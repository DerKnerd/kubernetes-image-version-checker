package quay

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"kubernetes-image-version-checker/config"
	"kubernetes-image-version-checker/container/imageCache"
	"net/http"
	"strings"
)

func Check(image string, registry config.Registry, logf func(message string, data ...interface{})) (*config.TagList, error) {
	tags := new(config.TagList)
	if tagList := imageCache.Check(image, logf); tagList != nil {
		return tagList, nil
	}

	logf("Get image version from quay.io")
	logf("Get all tags for image %s", image)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s/api/v1/repository/%s", registry.Host, strings.TrimPrefix(image, registry.Host+"/")), nil)
	if err != nil {
		return nil, err
	}
	if registry.Token != "" {
		req.Header.Set("Authentication", fmt.Sprintf("Bearer %s", registry.Token))
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get tags")
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body %s", err.Error())
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
	imageCache.Add(image, tags)

	return tags, nil
}
