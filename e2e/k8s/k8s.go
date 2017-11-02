package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Client returns an incluster k8s clientset.
func Client() (*kubernetes.Clientset, error) {
	// Create the in-cluster config.
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// Create the clientset.
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}
