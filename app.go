package main

import (
	"context"
	"github.com/hashicorp/go-version"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"kubernetes-pod-version-checker/dockerApi"
	"kubernetes-pod-version-checker/mailing"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	var (
		config *rest.Config
		err    error
	)
	if os.Getenv("MODE") == "out" {
		log.Println("Started in out cluster mode")
		home := homedir.HomeDir()
		config, err = clientcmd.BuildConfigFromFlags("", filepath.Join(home, ".kube", "config"))
		if err != nil {
			log.Fatalln(err.Error())
		}
	} else {
		log.Println("Started in in cluster mode")
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatalln(err.Error())
		}
	}

	log.Println("Create clientset for configuration")
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalln(err.Error())
	}

	log.Println("Look for deployments on kubernetes cluster")
	deployments, err := clientset.AppsV1().Deployments(apiv1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Fatalln(err.Error())
	}

	log.Printf("Found %d deployments", len(deployments.Items))

	log.Println("Start check for docker images")
	for _, deployment := range deployments.Items {
		for _, container := range deployment.Spec.Template.Spec.Containers {
			log.Printf("Check for docker image %s", container.Image)
			imageAndVersion := strings.Split(container.Image, ":")

			image := imageAndVersion[0]
			image = strings.ReplaceAll(image, os.Getenv("CUSTOM_REGISTRY_HOST"), "")

			log.Println("Get image version from docker hub")
			tagList, err := dockerApi.GetVersions(image)
			if err != nil {
				log.Println(err.Error())
				continue
			}

			if len(imageAndVersion) < 2 || imageAndVersion[1] == "latest" {
				continue
			}
			ver := imageAndVersion[1]

			log.Printf("Found %d tags for image %s", len(tagList.Tags), tagList.Name)
			log.Printf("Use version %s as constraint version", ver)
			versionConstraint, _ := version.NewConstraint("> " + ver)

			versions := make([]*version.Version, 0)
			for _, raw := range tagList.Tags {
				v, _ := version.NewVersion(raw)
				if v != nil {
					versions = append(versions, v)
				}
			}

			usedVersion, err := version.NewVersion(ver)
			if err != nil {
				log.Println(err.Error())
				continue
			}

			sort.Sort(sort.Reverse(version.Collection(versions)))
			log.Printf("Latest version for %s is %s", tagList.Name, versions[0].String())

			for _, tag := range versions {
				if versionConstraint.Check(tag) && !tag.LessThanOrEqual(usedVersion) {
					log.Printf("Found newer version for image %s:%s, newer version is %s", tagList.Name, usedVersion.String(), tag.String())
					if err = mailing.SendMail(*usedVersion, *tag, image, deployment); err != nil {
						log.Printf("Failed to send message for image %s", deployment.Name)
						log.Println(err.Error())
					}
					break
				}
			}
		}
	}
}
