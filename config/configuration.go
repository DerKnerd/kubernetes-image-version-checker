package config

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"kubernetes-pod-version-checker/messaging/mailing"
	"kubernetes-pod-version-checker/messaging/telegram"
)

type Configuration struct {
	Registries       []Registry       `yaml:"registries"`
	IgnoreNamespaces []string         `yaml:"ignore_namespaces"`
	Mode             string           `yaml:"cluster_mode"`
	Mailer           *mailing.Mailer  `yaml:"mailer"`
	TelegramClient   *telegram.Client `yaml:"telegram,omitempty"`
}

func ParseConfig(path string) (*Configuration, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Configuration
	err = yaml.Unmarshal(data, &config)

	return &config, err
}
