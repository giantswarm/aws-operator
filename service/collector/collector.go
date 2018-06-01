package collector

import (
	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
)

type Config struct {
	Logger    micrologger.Logger
	G8sClient versioned.Interface

	AwsConfig        awsutil.Config
	InstallationName string
}

type Collector struct {
	logger    micrologger.Logger
	g8sClient versioned.Interface

	awsClients       awsutil.Clients
	installationName string
}

func New(config Config) (*Collector, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	var emptyAwsConfig awsutil.Config
	if config.AwsConfig == emptyAwsConfig {
		return nil, microerror.Maskf(invalidConfigError, "%T.AwsConfig must not be empty", config)
	}
	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationName must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}

	awsClients := awsutil.NewClients(config.AwsConfig)

	c := &Collector{
		logger:    config.Logger,
		g8sClient: config.G8sClient,

		awsClients:       awsClients,
		installationName: config.InstallationName,
	}

	return c, nil
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- vpcsDesc
	ch <- clustersDesc
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.logger.Log("level", "debug", "message", "collecting metrics")

	c.collectVPCs(ch)

	c.collectClusterInfo(ch)

	c.logger.Log("level", "debug", "message", "finished collecting metrics")
}
