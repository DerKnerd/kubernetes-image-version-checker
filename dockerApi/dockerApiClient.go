package dockerApi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type authentication struct {
	Token string `json:"token"`
}

type TagList struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

var CheckedImages = map[string]*TagList{}

func GetVersions(image string) (*TagList, error) {
	log.Printf("Check if %s is in cache", image)
	if CheckedImages[image] != nil {
		return CheckedImages[image], nil
	}

	log.Println("Run authentication request")
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

	log.Println("Got api token")
	log.Printf("Get all tags for image %s", image)
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

	tags := new(TagList)
	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body")
	}

	err = json.Unmarshal(body, tags)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal body")
	}

	CheckedImages[image] = tags

	return tags, nil
}
