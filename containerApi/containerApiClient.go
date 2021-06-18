package containerApi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type authentication struct {
	Token string `json:"token"`
}

type TagList struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

var CheckedImages = map[string]*TagList{}

const quayIoPrefix = "quay.io"

func GetVersions(image string, logf func(message string, data ...interface{})) (*TagList, error) {
	tags := new(TagList)
	logf("Check if %s is in cache", image)
	if CheckedImages[image] != nil {
		return CheckedImages[image], nil
	}

	if strings.HasPrefix(image, quayIoPrefix) {
		logf("Get image version from quay.io")
		logf("Get all tags for image %s", image)

		response, err := http.DefaultClient.Get("https://quay.io/api/v1/repository/" + strings.TrimPrefix(image, "quay.io/"))
		if err != nil {
			return nil, err
		}

		if response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to get tags")
		}

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read body")
		}

		type containerImage struct {
			Name string                            `json:"name"`
			Tags map[string]map[string]interface{} `json:"tags"`
		}

		var cImage containerImage

		err = json.Unmarshal(body, &cImage)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal body %s", err.Error())
		}

		for k := range cImage.Tags {
			tags.Tags = append(tags.Tags, k)
		}
		tags.Name = cImage.Name
	} else {
		logf("Get image version from docker hub")
		logf("Run authentication request")
		authResponse, err := http.DefaultClient.Get("https://auth.docker.io/token?service=registry.docker.io&scope=repository:" + image + ":pull")
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
		req, err := http.NewRequest(http.MethodGet, "https://registry-1.docker.io/v2/"+image+"/tags/list", nil)
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
	}

	CheckedImages[image] = tags

	return tags, nil
}
