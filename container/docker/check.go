package docker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"kubernetes-pod-version-checker/container"
	"kubernetes-pod-version-checker/container/imageCache"
	"net/http"
)

type authentication struct {
	Token string `json:"token"`
}

func Check(image string, registry container.Registry, logf func(message string, data ...interface{})) (*container.TagList, error) {
	tags := new(container.TagList)
	if tagList := imageCache.Check(image, logf); tagList != nil {
		return tagList, nil
	}

	logf("Get image version from docker hub")
	logf("Run authentication request")
	authResponse, err := http.DefaultClient.Get(fmt.Sprintf("https://%s/token?service=registry.docker.io&scope=repository:%s:pull", registry.AuthHost, image))
	if err != nil {
		return nil, err
	}
	if authResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to login")
	}

	auth := new(authentication)
	body, err := ioutil.ReadAll(authResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body")
	}

	err = json.Unmarshal(body, auth)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal body")
	}

	logf("Got api token")
	logf("Get all tags for image %s", image)
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s/v2/%s/tags/list", registry.Host, image), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request")
	}

	req.Header.Set("Authorization", "Bearer "+auth.Token)
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get tags")
	}

	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body")
	}

	err = json.Unmarshal(body, tags)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal body %s", err.Error())
	}

	imageCache.Add(image, tags)

	return tags, nil
}
