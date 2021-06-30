package container

import (
	"github.com/hashicorp/go-version"
	"sort"
	"strings"
)

func CheckVersions(currentVersion string, tagList *TagList, logf func(message string, data ...interface{})) (tagVersion string, outdated bool) {
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

	usedVersion, err := version.NewVersion(trimmedVersion)
	if err != nil {
		logf(err.Error())
		return
	}

	sort.Sort(sort.Reverse(version.Collection(versions)))
	tag := versions[0]
	logf("Latest version for %s is %s", tagList.Name, versions[0].String())
	if versionConstraint.Check(tag) && !tag.LessThanOrEqual(usedVersion) {
		logf("Found newer version for image %s:%s, newer version is %s", tagList.Name, usedVersion.String(), tag.String())
		tagVersion = tag.Original()
		if len(versionAndSuffixSplit) > 1 {
			tagVersion += "-" + strings.Join(versionAndSuffixSplit[1:], "")
		}
		outdated = true
	}

	return
}
