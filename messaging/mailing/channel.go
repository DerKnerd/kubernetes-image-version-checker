package mailing

import (
	"fmt"
	"kubernetes-pod-version-checker/messaging"
	"strconv"
)

func Processor(mailer *Mailer, mailChan chan messaging.Message, logChan chan string) {
	for message := range mailChan {
		if err := mailer.SendMail(message); err != nil {
			logf := func(logMessage string, data ...interface{}) {
				logChan <- fmt.Sprintf("CPU "+strconv.Itoa(message.Cpu)+": "+logMessage, data...)
			}
			logf("Failed to send mail message for image %s", message.ParentName)
			logf(err.Error())
		}
	}
}
