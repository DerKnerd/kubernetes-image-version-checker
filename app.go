package main

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-version"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"kubernetes-pod-version-checker/containerApi"
	"kubernetes-pod-version-checker/mailing"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

type containerData struct {
	container  apiv1.Container
	parentName string
	entityType string
}

func main() {
	var (
		wg     = &sync.WaitGroup{}
		config *rest.Config
		err    error
	)
	if os.Getenv("MODE") == "out" {
		log.Println("MAIN:  Started in out cluster mode")
		home := homedir.HomeDir()
		config, err = clientcmd.BuildConfigFromFlags("", filepath.Join(home, ".kube", "config"))
		if err != nil {
			log.Fatalln(err.Error())
		}
	} else {
		log.Println("MAIN:  Started in in cluster mode")
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatalln(err.Error())
		}
	}

	ignoreNamespacesVar := os.Getenv("IGNORE_NAMESPACES")
	ignoreNamespaces := strings.Split(ignoreNamespacesVar, ",")

	log.Printf("MAIN:  Found %d CPUs", runtime.NumCPU())
	wg.Add(runtime.NumCPU())

	log.Println("MAIN:  Create clientset for configuration")
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalln(err.Error())
	}

	containers := make([]containerData, 0)

	log.Println("MAIN:  Look for deployments on kubernetes cluster")
	deployments, err := clientset.AppsV1().Deployments(apiv1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Println(err.Error())
	} else {
		log.Printf("MAIN:  Found %d deployments", len(deployments.Items))

		log.Println("MAIN:  Start check for deployment updates")
		for _, deployment := range deployments.Items {
			if contains(ignoreNamespaces, deployment.GetNamespace()) {
				continue
			}

			for _, container := range deployment.Spec.Template.Spec.Containers {
				containers = append(containers, containerData{
					container:  container,
					parentName: deployment.GetName(),
					entityType: "deployment",
				})
			}
		}
	}

	log.Println("MAIN:  Look for daemon sets on kubernetes cluster")
	daemonSets, err := clientset.AppsV1().DaemonSets(apiv1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Println(err.Error())
	} else {
		log.Printf("MAIN:  Found %d daemon sets", len(daemonSets.Items))

		log.Println("MAIN:  Start check for daemon set updates")
		for _, daemonSet := range daemonSets.Items {
			if contains(ignoreNamespaces, daemonSet.GetNamespace()) {
				continue
			}

			for _, container := range daemonSet.Spec.Template.Spec.Containers {
				containers = append(containers, containerData{
					container:  container,
					parentName: daemonSet.GetName(),
					entityType: "daemon set",
				})
			}
		}
	}

	log.Println("MAIN:  Look for stateful sets on kubernetes cluster")
	statefulSets, err := clientset.AppsV1().StatefulSets(apiv1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Println(err.Error())
	} else {
		log.Printf("MAIN:  Found %d stateful sets", len(statefulSets.Items))

		log.Println("MAIN:  Start check for stateful set updates")
		for _, statefulSet := range statefulSets.Items {
			if contains(ignoreNamespaces, statefulSet.GetNamespace()) {
				continue
			}

			for _, container := range statefulSet.Spec.Template.Spec.Containers {
				containers = append(containers, containerData{
					container:  container,
					parentName: statefulSet.GetName(),
					entityType: "stateful set",
				})
			}
		}
	}

	log.Println("MAIN:  Look for cron jobs on kubernetes cluster")
	cronJobs, err := clientset.BatchV1().CronJobs(apiv1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Println(err.Error())
	} else {
		log.Printf("MAIN:  Found %d cron job", len(cronJobs.Items))

		log.Println("MAIN:  Start check for cron job updates")
		for _, cronJob := range cronJobs.Items {
			if contains(ignoreNamespaces, cronJob.GetNamespace()) {
				continue
			}

			for _, container := range cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers {
				containers = append(containers, containerData{
					container:  container,
					parentName: cronJob.GetName(),
					entityType: "cron job",
				})
			}
		}
	}

	containerChan := make(chan containerData)
	logChan := make(chan string)

	for i := 0; i < runtime.NumCPU(); i++ {
		log.Printf("MAIN:  Start process on cpu %d", i)
		go processContainer(wg, containerChan, logChan, i)
	}

	go func(logChan chan string) {
		for logEntry := range logChan {
			log.Println(logEntry)
		}
	}(logChan)

	for _, container := range containers {
		containerChan <- container
	}

	close(containerChan)

	wg.Wait()
	close(logChan)
}

func processContainer(wg *sync.WaitGroup, containerChan chan containerData, logChan chan string, idx int) {
	for c := range containerChan {
		logf := func(message string, data ...interface{}) {
			logChan <- fmt.Sprintf("CPU "+strconv.Itoa(idx)+": "+message, data...)
		}
		checkContainerForUpdates(c.container, c.parentName, c.entityType, logf)
	}
	logChan <- fmt.Sprintf("CPU %d: Process ended", idx)
	wg.Done()
}

func checkContainerForUpdates(container apiv1.Container, parentName string, entityType string, logf func(message string, data ...interface{})) {
	logf("Check for container image %s", container.Image)
	imageAndVersion := strings.Split(container.Image, ":")

	image := imageAndVersion[0]
	image = strings.ReplaceAll(image, os.Getenv("CUSTOM_REGISTRY_HOST"), "")

	tagList, err := containerApi.GetVersions(image, logf)
	if err != nil {
		logf(err.Error())
		return
	}

	if len(imageAndVersion) < 2 || imageAndVersion[1] == "latest" {
		return
	}
	ver := imageAndVersion[1]

	versionAndSuffixSplit := strings.Split(ver, "-")
	trimmedVersion := versionAndSuffixSplit[0]

	logf("Found %d tags for image %s", len(tagList.Tags), tagList.Name)
	logf("Use version %s as constraint version", ver)
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
		tagVersion := tag.Original()
		if len(versionAndSuffixSplit) > 1 {
			tagVersion += "-" + strings.Join(versionAndSuffixSplit[1:], "")
		}
		if err = mailing.SendMail(ver, tagVersion, image, parentName, entityType); err != nil {
			logf("Failed to send message for image %s", parentName)
			logf(err.Error())
		}
	}

	return
}
