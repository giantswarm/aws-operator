package k8s

import (
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/auth"
)

var (
	filename string = fmt.Sprintf("%s/.kube/config", os.Getenv("HOME"))
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

	info, err := auth.LoadFromFile(filename)
	if err != nil {
		return nil, err
	}

	rawClientConfig = info.MergeWithConfig()

	client, err := kubernetes.NewForConfig(rawClientConfig)
	if err != nil {
		return nil, err
	}

	return client, nil
}
