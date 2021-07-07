package worker

import (
	"kubernetes-pod-version-checker/config"
	"kubernetes-pod-version-checker/container"
	"kubernetes-pod-version-checker/kubernetes"
	"kubernetes-pod-version-checker/logging"
	"kubernetes-pod-version-checker/messaging"
	mailing "kubernetes-pod-version-checker/messaging/mailing"
	"kubernetes-pod-version-checker/messaging/telegram"
	"kubernetes-pod-version-checker/worker/configFile"
	"log"
	"runtime"
	"sync"
)

func ExecuteWithConfigFile(path string) error {
	log.Printf("MAIN:  Parse config %s", path)
	configuration, err := config.ParseConfig(path)
	if err != nil {
		return err
	}

	var (
		wg = &sync.WaitGroup{}
	)

	log.Printf("MAIN:  Found %d CPUs", runtime.NumCPU())
	wg.Add(runtime.NumCPU())

	log.Println("MAIN:  Create clientset for configuration")
	clientset, err := kubernetes.GetClientSet(configuration.Mode)
	if err != nil {
		return err
	}

	containerChan := make(chan container.Details)
	logChan := make(chan string)
	messagingChan := make(chan messaging.Message)
	mailChan := make(chan messaging.Message)
	telegramChan := make(chan messaging.Message)

	for i := 0; i < runtime.NumCPU(); i++ {
		log.Printf("MAIN:  Start process on cpu %d", i)
		go container.ProcessContainer(wg, containerChan, logChan, messagingChan, i, configFile.CheckContainerForUpdates(configuration.Registries))
	}

	go func(messagingChan chan messaging.Message, mailChan chan messaging.Message, telegramChan chan messaging.Message) {
		for m := range messagingChan {
			mailChan <- m
			if configuration.TelegramClient != nil {
				telegramChan <- m
			}
		}
	}(messagingChan, mailChan, telegramChan)
	go mailing.Processor(&configuration.Mailer, mailChan, logChan)
	if configuration.TelegramClient != nil {
		go telegram.Processor(configuration.TelegramClient, telegramChan, logChan)
	}
	go logging.Processor(logChan)

	for _, c := range kubernetes.ExtractContainer(clientset, configuration.IgnoreNamespaces) {
		containerChan <- c
	}

	close(containerChan)

	wg.Wait()
	close(telegramChan)
	close(mailChan)
	close(messagingChan)
	close(logChan)

	return nil
}
