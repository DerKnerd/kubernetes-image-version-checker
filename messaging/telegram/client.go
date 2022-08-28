package telegram

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"kubernetes-image-version-checker/messaging"
)

type Client struct {
	ChannelId int64  `yaml:"channel_id"`
	BotToken  string `yaml:"bot_token"`
}

func (receiver Client) SendMessage(message messaging.Message) error {
	api, err := tgbotapi.NewBotAPI(receiver.BotToken)
	if err != nil {
		return err
	}

	text := fmt.Sprintf("%s %s: %s:%s -> %s:%s", message.EntityType, message.ParentName, message.Image, message.UsedVersion, message.Image, message.LatestVersion)
	msg := tgbotapi.NewMessage(receiver.ChannelId, text)
	_, err = api.Send(msg)

	return err
}
