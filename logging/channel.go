package logging

import "log"

func Processor(logChan chan string) {
	for logEntry := range logChan {
		log.Println(logEntry)
	}
}
