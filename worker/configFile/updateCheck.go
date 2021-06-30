package configFile

import (
	apiv1 "k8s.io/api/core/v1"
	"kubernetes-pod-version-checker/container"
	"kubernetes-pod-version-checker/container/docker"
	"kubernetes-pod-version-checker/container/quay"
	"kubernetes-pod-version-checker/mailing"
	"strings"
)

func findRegistry(registries []container.Registry, image string) (*container.Registry, bool) {
	for _, registry := range registries {
		if strings.HasPrefix(image, registry.Host) {
			return &registry, true
		}
	}

	return nil, false
}

func CheckContainerForUpdates(registries []container.Registry) func(apiv1.Container, string, string, chan mailing.Message, int, func(string, ...interface{})) {
	return func(c apiv1.Container, parentName string, entityType string, mailChan chan mailing.Message, cpu int, logf func(message string, data ...interface{})) {
		logf("Check for container image %s", c.Image)
		imageAndVersion := strings.Split(c.Image, ":")

		image := imageAndVersion[0]
		registry, found := findRegistry(registries, image)
		if !found {
			tmpRegistry := container.NewDockerHub()
			registry = &tmpRegistry
		}

		var (
			tagList *container.TagList
			err     error
		)

		if registry.Type == container.RegistryTypeQuay {
			tagList, err = quay.Check(image, *registry, logf)
			if err != nil {
				logf(err.Error())
				return
			}
		} else if registry.Type == container.RegistryTypeDockerHub {
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

		tagVersion, outdated := container.CheckVersions(currentVersion, tagList, logf)
		if outdated {
			mailChan <- mailing.Message{
				UsedVersion:   currentVersion,
				LatestVersion: tagVersion,
				Image:         image,
				ParentName:    parentName,
				EntityType:    entityType,
				Cpu:           cpu,
			}
		}
	}
}
