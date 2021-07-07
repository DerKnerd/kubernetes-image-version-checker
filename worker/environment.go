package worker

import (
	"kubernetes-pod-version-checker/config"
	"kubernetes-pod-version-checker/container"
	"kubernetes-pod-version-checker/kubernetes"
	"kubernetes-pod-version-checker/logging"
	"kubernetes-pod-version-checker/messaging/mailing"
	"kubernetes-pod-version-checker/worker/environment"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
)

func ExecuteWithEnvironment() error {
	var (
		wg  = &sync.WaitGroup{}
		err error
	)

	ignoreNamespacesVar := os.Getenv("IGNORE_NAMESPACES")
	ignoreNamespaces := strings.Split(ignoreNamespacesVar, ",")

	log.Printf("MAIN:  Found %d CPUs", runtime.NumCPU())
	wg.Add(runtime.NumCPU())

	log.Println("MAIN:  Create clientset for configuration")
	clientset, err := kubernetes.GetClientSet(os.Getenv("MODE"))
	if err != nil {
		return err
	}

	containerChan := make(chan config.Details)
	logChan := make(chan string)

	configuration := &config.Configuration{Mailer: mailing.New(
		[]string{os.Getenv("MAILING_TO")},
		os.Getenv("MAILING_FROM"),
		os.Getenv("MAILING_USERNAME"),
		os.Getenv("MAILING_PASSWORD"),
		os.Getenv("MAILING_HOST"),
		os.Getenv("MAILING_PORT"),
	)}
	for i := 0; i < runtime.NumCPU(); i++ {
		log.Printf("MAIN:  Start process on cpu %d", i)
		go container.ProcessContainer(wg, configuration, containerChan, logChan, i, environment.CheckContainerForUpdates)
	}

	go logging.Processor(logChan)

	for _, c := range kubernetes.ExtractContainer(clientset, ignoreNamespaces) {
		containerChan <- c
	}

	close(containerChan)

	wg.Wait()
	close(logChan)

	return nil
}
