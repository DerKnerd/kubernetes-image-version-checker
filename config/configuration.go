package config

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"kubernetes-pod-version-checker/container"
	"kubernetes-pod-version-checker/mailing"
)

type Configuration struct {
	Registries       []container.Registry `yaml:"registries"`
	IgnoreNamespaces []string             `yaml:"ignore_namespaces"`
	Mode             string               `yaml:"cluster_mode"`
	Mailer           mailing.Mailer       `yaml:"mailer"`
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
