package container

import (
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	"kubernetes-pod-version-checker/messaging"
	"strconv"
	"sync"
)

func ProcessContainer(wg *sync.WaitGroup, containerChan chan Details, logChan chan string, mailChan chan messaging.Message, cpu int, checkContainerForUpdates func(apiv1.Container, string, string, chan messaging.Message, int, func(message string, data ...interface{}))) {
	for c := range containerChan {
		logf := func(message string, data ...interface{}) {
			logChan <- fmt.Sprintf("CPU "+strconv.Itoa(cpu)+": "+message, data...)
		}
		checkContainerForUpdates(c.Container, c.ParentName, c.EntityType, mailChan, cpu, logf)
	}
	logChan <- fmt.Sprintf("CPU %d: Process ended", cpu)
	wg.Done()
}
