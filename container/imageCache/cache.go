package imageCache

import (
	"kubernetes-image-version-checker/config"
	"sync"
)

var CheckedImages = map[string]*config.TagList{}
var checkImageMutex = sync.Mutex{}

func Check(image string, logf func(message string, data ...interface{})) *config.TagList {
	logf("Check if %s is in cache", image)
	checkImageMutex.Lock()
	if CheckedImages[image] != nil {
		defer checkImageMutex.Unlock()
		return CheckedImages[image]
	}
	checkImageMutex.Unlock()

	return nil
}

func Add(image string, tags *config.TagList) {
	checkImageMutex.Lock()
	CheckedImages[image] = tags
	checkImageMutex.Unlock()
}
