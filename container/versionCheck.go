package container

import (
	"github.com/hashicorp/go-version"
	"kubernetes-pod-version-checker/config"
	"regexp"
	"sort"
	"strings"
)

func CheckVersions(image string, currentVersion string, tagList *config.TagList, logf func(message string, data ...interface{})) (tagVersion string, outdated bool) {
	if currentVersion == "latest" || currentVersion == "" {
		return
	}
	outdated = false

	versionAndSuffixSplit := strings.Split(currentVersion, "-")
	trimmedVersion := versionAndSuffixSplit[0]

	logf("Found %d tags for image %s", len(tagList.Tags), tagList.Name)
	logf("Use version %s as constraint version", currentVersion)
	versionConstraint, _ := version.NewConstraint("> " + trimmedVersion)

	versions := make([]*version.Version, 0)
	for _, raw := range tagList.Tags {
		v, _ := version.NewVersion(raw)
		if v != nil {
			versions = append(versions, v)
		}
	}

	filteredVersion := versions
	configuration := config.GetCurrentConfig()
	if configuration != nil {
		if configuration.ImageFormatConstraints != nil {
			containsImage := false
			for key, _ := range configuration.ImageFormatConstraints {
				if key == image {
					containsImage = true
					break
				}
			}

			if containsImage {
				constraint := configuration.ImageFormatConstraints[image]
				regex, err := regexp.Compile(constraint)
				if err == nil {
					filteredVersion = make([]*version.Version, 0)
					for _, v := range versions {
						if regex.MatchString(v.Original()) {
							filteredVersion = append(filteredVersion, v)
						}
					}
				}
			}
		}
	}

	usedVersion, err := version.NewVersion(trimmedVersion)
	if err != nil {
		logf(err.Error())
		return
	}

	if len(filteredVersion) > 0 {
		sort.Sort(sort.Reverse(version.Collection(filteredVersion)))
		tag := filteredVersion[0]
		logf("Latest version for %s is %s", tagList.Name, filteredVersion[0].String())
		checkVersion := tag.Core()
		if versionConstraint.Check(checkVersion) && !checkVersion.LessThanOrEqual(usedVersion) {
			logf("Found newer version for image %s:%s, newer version is %s", tagList.Name, usedVersion.String(), checkVersion.String())
			tagVersion = checkVersion.String()
			if len(versionAndSuffixSplit) > 1 {
				tagVersion += "-" + strings.Join(versionAndSuffixSplit[1:], "")
			}
			outdated = true
		}
	}

	return
}
