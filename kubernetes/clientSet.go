package kubernetes

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"path/filepath"
)

func GetClientSet(mode string) (*kubernetes.Clientset, error) {
	var (
		config *rest.Config
		err    error
	)
	if mode == "out" {
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

	return kubernetes.NewForConfig(config)
}
