package configFile

import (
	apiv1 "k8s.io/api/core/v1"
	"kubernetes-pod-version-checker/config"
	"kubernetes-pod-version-checker/container"
	"kubernetes-pod-version-checker/container/docker"
	"kubernetes-pod-version-checker/container/quay"
	"kubernetes-pod-version-checker/messaging"
	"strings"
)

func findRegistry(registries []config.Registry, image string) (*config.Registry, bool) {
	for _, registry := range registries {
		if strings.HasPrefix(image, registry.Host) {
			return &registry, true
		}
	}

	return nil, false
}

func CheckContainerForUpdates(registries []config.Registry) func(*config.Configuration, apiv1.Container, string, string, int, func(string, ...interface{})) {
	return func(configuration *config.Configuration, c apiv1.Container, parentName string, entityType string, cpu int, logf func(message string, data ...interface{})) {
		logf("Check for container image %s", c.Image)
		imageAndVersion := strings.Split(c.Image, ":")

		image := imageAndVersion[0]
		registry, found := findRegistry(registries, image)
		if !found {
			tmpRegistry := config.NewDockerHub()
			registry = &tmpRegistry
		}

		var (
			tagList *config.TagList
			err     error
		)

		if registry.Type == config.RegistryTypeQuay {
			tagList, err = quay.Check(image, *registry, logf)
			if err != nil {
				logf(err.Error())
				return
			}
		} else if registry.Type == config.RegistryTypeDockerHub {
			tagList, err = docker.Check(image, *registry, logf)
			if err != nil {
				logf(err.Error())
				return
			}
		} else {
			logf("invalid registry type %s", registry.Type)
			return
		}

		currentVersion := "latest"
		if len(imageAndVersion) >= 2 {
			currentVersion = imageAndVersion[1]
		}

		tagVersion, outdated := container.CheckVersions(image, currentVersion, tagList, logf)
		if outdated {
			message := messaging.Message{
				UsedVersion:   currentVersion,
				LatestVersion: tagVersion,
				Image:         image,
				ParentName:    parentName,
				EntityType:    entityType,
				Cpu:           cpu,
			}
			if configuration.Mailer != nil {
				err = configuration.Mailer.SendMail(message)
				if err != nil {
					logf(err.Error())
				}
			}
			if configuration.TelegramClient != nil {
				err = configuration.TelegramClient.SendMessage(message)
				if err != nil {
					logf(err.Error())
				}
			}
		}
	}
}
