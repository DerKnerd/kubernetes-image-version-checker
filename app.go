package main

import (
	"flag"
	"kubernetes-pod-version-checker/worker"
	"log"
)

var (
	configTypeFlag = flag.String("config-mode", "file", "Set env for environment configuration and, file for config file based configuration")
	configFileFlag = flag.String("config-file", "", "Sets the config file path")
)

func main() {
	flag.Parse()

	if *configTypeFlag == "env" {
		err := worker.ExecuteWithEnvironment()
		if err != nil {
			log.Fatalln(err)
		}
	} else if *configTypeFlag == "file" {
		err := worker.ExecuteWithConfigFile(*configFileFlag)
		if err != nil {
			log.Fatalln(err)
		}
	}
}
