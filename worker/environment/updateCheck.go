package environment

import (
	apiv1 "k8s.io/api/core/v1"
	container "kubernetes-pod-version-checker/container"
	"kubernetes-pod-version-checker/container/docker"
	"kubernetes-pod-version-checker/container/quay"
	"kubernetes-pod-version-checker/mailing"
	"os"
	"strings"
)

func CheckContainerForUpdates(c apiv1.Container, parentName string, entityType string, mailChan chan mailing.Message, cpu int, logf func(message string, data ...interface{})) {
	logf("Check for container image %s", c.Image)
	imageAndVersion := strings.Split(c.Image, ":")

	image := imageAndVersion[0]
	image = strings.ReplaceAll(image, os.Getenv("CUSTOM_REGISTRY_HOST"), "")

	var (
		tagList *container.TagList
		err     error
	)

	if strings.HasPrefix(image, "quay.io") {
		quayRegistry := container.NewQuayIo()
		tagList, err = quay.Check(image, quayRegistry, logf)
		if err != nil {
			logf(err.Error())
			return
		}
	} else {
		dockerRegistry := container.NewDockerHub()
		tagList, err = docker.Check(image, dockerRegistry, logf)
		if err != nil {
			logf(err.Error())
			return
		}
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