package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func NewClient(host, username, password, bearerToken string, insecure bool) (kubernetes.Interface, error) {
	cfg := &rest.Config{
		Host:        host,
		QPS:         100,
		Burst:       100,
		Username:    username,
		Password:    password,
		BearerToken: bearerToken,
		Insecure:    insecure,
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return client, nil
}
