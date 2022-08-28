package kubernetes

import (
	"context"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"kubernetes-image-version-checker/config"
	"kubernetes-image-version-checker/utils"
	"log"
)

func addContainer(containers []apiv1.Container, entityType string, name string) []config.Details {
	var result []config.Details
	for _, c := range containers {
		result = append(result, config.Details{
			Container:  c,
			ParentName: name,
			EntityType: entityType,
		})
	}

	return result
}

func ExtractContainer(clientset *kubernetes.Clientset, ignoreNamespaces []string) []config.Details {
	containers := make([]config.Details, 0)

	log.Println("MAIN:  Look for deployments on kubernetes cluster")
	deployments, err := clientset.AppsV1().Deployments(apiv1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Println(err.Error())
	} else {
		log.Printf("MAIN:  Found %d deployments", len(deployments.Items))

		log.Println("MAIN:  Start check for deployment updates")
		for _, deployment := range deployments.Items {
			if utils.ContainsString(ignoreNamespaces, deployment.GetNamespace()) {
				continue
			}
			containers = append(containers, addContainer(deployment.Spec.Template.Spec.Containers, "deployment", deployment.GetName())...)
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
			if utils.ContainsString(ignoreNamespaces, daemonSet.GetNamespace()) {
				continue
			}

			containers = append(containers, addContainer(daemonSet.Spec.Template.Spec.Containers, "daemon set", daemonSet.GetName())...)
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
			if utils.ContainsString(ignoreNamespaces, statefulSet.GetNamespace()) {
				continue
			}

			containers = append(containers, addContainer(statefulSet.Spec.Template.Spec.Containers, "stateful set", statefulSet.GetName())...)
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
			if utils.ContainsString(ignoreNamespaces, cronJob.GetNamespace()) {
				continue
			}

			containers = append(containers, addContainer(cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers, "cron job", cronJob.GetName())...)
		}
	}

	return containers
}
