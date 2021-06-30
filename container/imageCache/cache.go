package imageCache

import (
	"kubernetes-pod-version-checker/container"
	"sync"
)

var CheckedImages = map[string]*container.TagList{}
var checkImageMutex = sync.Mutex{}

func Check(image string, logf func(message string, data ...interface{})) *container.TagList {
	logf("Check if %s is in cache", image)
	checkImageMutex.Lock()
	if CheckedImages[image] != nil {
		defer checkImageMutex.Unlock()
		return CheckedImages[image]
	}
	checkImageMutex.Unlock()

	return nil
}

func Add(image string, tags *container.TagList) {
	checkImageMutex.Lock()
	CheckedImages[image] = tags
	checkImageMutex.Unlock()
}
