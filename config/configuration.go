package config

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"kubernetes-image-version-checker/messaging/mailing"
	"kubernetes-image-version-checker/messaging/telegram"
)

type Configuration struct {
	Registries             []Registry        `yaml:"registries"`
	IgnoreNamespaces       []string          `yaml:"ignore_namespaces"`
	Mode                   string            `yaml:"cluster_mode"`
	Mailer                 *mailing.Mailer   `yaml:"mailer"`
	TelegramClient         *telegram.Client  `yaml:"telegram,omitempty"`
	ImageFormatConstraints map[string]string `yaml:"image_format_constraints"`
}

var configuration *Configuration

func GetCurrentConfig() *Configuration {
	return configuration
}

func ParseConfig(path string) (*Configuration, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Configuration
	err = yaml.Unmarshal(data, &config)

	configuration = &config

	return &config, err
}
