package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type TLSClientConfig struct {
	// Files containing keys/certificates.
	CertFile string
	KeyFile  string
	CAFile   string
}

type Config struct {
	InCluster   bool
	Host        string
	Username    string
	Password    string
	BearerToken string
	TLSClientConfig
	inClusterConfigProvider func() (*rest.Config, error)
}

func getInClusterConfig(config Config) (*rest.Config, error) {
	if config.inClusterConfigProvider == nil {
		config.inClusterConfigProvider = rest.InClusterConfig
	}

	return config.inClusterConfigProvider()
}

func newRawClientConfig(config Config) *rest.Config {
	tlsClientConfig := rest.TLSClientConfig{
		CertFile: config.CertFile,
		KeyFile:  config.KeyFile,
		CAFile:   config.CAFile,
	}
	rawClientConfig := &rest.Config{
		Host:            config.Host,
		QPS:             100,
		Burst:           100,
		Username:        config.Username,
		Password:        config.Password,
		BearerToken:     config.BearerToken,
		TLSClientConfig: tlsClientConfig,
	}

	return rawClientConfig
}

func getRawClientConfig(config Config) (*rest.Config, error) {
	var rawClientConfig *rest.Config
	var err error

	if config.InCluster {
		rawClientConfig, err = getInClusterConfig(config)
		if err != nil {
			return nil, err
		}
	} else {
		rawClientConfig = newRawClientConfig(config)
	}

	return rawClientConfig, nil
}

func NewClient(config Config) (kubernetes.Interface, error) {
	rawClientConfig, err := getRawClientConfig(config)
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(rawClientConfig)
	if err != nil {
		return nil, err
	}

	return client, nil
}
