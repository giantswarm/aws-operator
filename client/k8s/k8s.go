package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Config struct {
	Host        string
	Username    string
	Password    string
	BearerToken string
	Insecure    bool
}

func NewClient(config Config) (kubernetes.Interface, error) {
	rawClientConfig := &rest.Config{
		Host:        config.Host,
		QPS:         100,
		Burst:       100,
		Username:    config.Username,
		Password:    config.Password,
		BearerToken: config.BearerToken,
		Insecure:    config.Insecure,
	}
	client, err := kubernetes.NewForConfig(rawClientConfig)
	if err != nil {
		return nil, err
	}
	return client, nil
}
