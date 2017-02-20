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

	// PEM-encoded keys.
	CertData string
	KeyData  string
	CAData   string
}

type Config struct {
	Host        string
	Username    string
	Password    string
	BearerToken string
	Insecure    bool
	TLSClientConfig
	inClusterConfigProvider func() (*rest.Config, error)
}

func getRawClientConfig(config Config) *rest.Config {
	if config.inClusterConfigProvider == nil {
		config.inClusterConfigProvider = rest.InClusterConfig
	}

	rawClientConfig, err := config.inClusterConfigProvider()
	if err != nil {
		tlsClientConfig := rest.TLSClientConfig{
			CertFile: config.CertFile,
			KeyFile:  config.KeyFile,
			CAFile:   config.CAFile,
		}
		if config.CertData != "" {
			tlsClientConfig.CertData = []byte(config.CertData)
		}
		if config.KeyData != "" {
			tlsClientConfig.KeyData = []byte(config.KeyData)
		}
		if config.CAData != "" {
			tlsClientConfig.CAData = []byte(config.CAData)
		}
		rawClientConfig = &rest.Config{
			Host:            config.Host,
			QPS:             100,
			Burst:           100,
			Username:        config.Username,
			Password:        config.Password,
			BearerToken:     config.BearerToken,
			Insecure:        config.Insecure,
			TLSClientConfig: tlsClientConfig,
		}
	}

	return rawClientConfig
}

func NewClient(config Config) (kubernetes.Interface, error) {
	rawClientConfig := getRawClientConfig(config)

	client, err := kubernetes.NewForConfig(rawClientConfig)
	if err != nil {
		return nil, err
	}
	return client, nil
}
