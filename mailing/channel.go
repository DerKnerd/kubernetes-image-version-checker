package mailing

import (
	"fmt"
	"strconv"
)

func Processor(mailer *Mailer, mailChan chan Message, logChan chan string) {
	for message := range mailChan {
		if err := mailer.SendMail(message); err != nil {
			logf := func(logMessage string, data ...interface{}) {
				logChan <- fmt.Sprintf("CPU "+strconv.Itoa(message.Cpu)+": "+logMessage, data...)
			}
			logf("Failed to send message for image %s", message.ParentName)
			logf(err.Error())
		}
	}
}
