package k8s

import (
	"github.com/ericchiang/k8s"
	microerror "github.com/giantswarm/microkit/error"
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
	Insecure    bool
	TLSClientConfig
}

func newRawClientConfig(config Config) *k8s.Config {
	rawClientConfig := &k8s.Config{
		Preferences: k8s.Preferences{},
		Clusters: []k8s.NamedCluster{k8s.NamedCluster{
			Name: "",
			Cluster: k8s.Cluster{
				Server:               config.Host,
				CertificateAuthority: config.CAFile,
			},
		}},
		AuthInfos: []k8s.NamedAuthInfo{k8s.NamedAuthInfo{
			Name: "",
			AuthInfo: k8s.AuthInfo{
				ClientCertificate: config.CertFile,
				ClientKey:         config.KeyFile,
				Token:             config.BearerToken,
				Username:          config.Username,
				Password:          config.Password,
			},
		}},
	}

	return rawClientConfig
}

func NewClient(config Config) (*k8s.Client, error) {
	if config.InCluster {
		client, err := k8s.NewInClusterClient()
		if err != nil {
			return nil, microerror.MaskAny(err)
		}

		return client, nil
	}

	rawClientConfig := newRawClientConfig(config)

	client, err := k8s.NewClient(rawClientConfig)
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	return client, nil
}
