package config

import (
	apiv1 "k8s.io/api/core/v1"
)

type TagList struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

type Details struct {
	Container  apiv1.Container
	ParentName string
	EntityType string
}

type Registry struct {
	Name     string `yaml:"name"`
	Host     string `yaml:"host"`
	AuthHost string `yaml:"auth_host"`
	Type     string `yaml:"type"`
	Username string `yaml:"username"`
	Token    string `yaml:"token"`
}

func NewDockerHub() Registry {
	return Registry{
		Name:     "DockerHub",
		Host:     "registry-1.docker.io",
		AuthHost: "auth.docker.io",
		Type:     RegistryTypeDockerHub,
	}
}

func NewQuayIo() Registry {
	return Registry{
		Name: "quay.io",
		Host: "quay.io",
		Type: RegistryTypeQuay,
	}
}

const RegistryTypeDockerHub = "dockerhub"
const RegistryTypeQuay = "quay"
