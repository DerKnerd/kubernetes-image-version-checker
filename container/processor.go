package container

import (
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	"kubernetes-pod-version-checker/config"
	"strconv"
	"sync"
)

func ProcessContainer(wg *sync.WaitGroup, config *config.Configuration, containerChan chan config.Details, logChan chan string, cpu int, checkContainerForUpdates func(*config.Configuration, apiv1.Container, string, string, int, func(message string, data ...interface{}))) {
	for c := range containerChan {
		logf := func(message string, data ...interface{}) {
			logChan <- fmt.Sprintf("CPU "+strconv.Itoa(cpu)+": "+message, data...)
		}
		checkContainerForUpdates(config, c.Container, c.ParentName, c.EntityType, cpu, logf)
	}
	logChan <- fmt.Sprintf("CPU %d: Process ended", cpu)
	wg.Done()
}
