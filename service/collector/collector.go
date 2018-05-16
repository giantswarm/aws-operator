package collector

import (
	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type Config struct {
	Logger micrologger.Logger

	AwsConfig        awsutil.Config
	InstallationName string
}

type Collector struct {
	logger micrologger.Logger

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

	awsClients := awsutil.NewClients(config.AwsConfig)

	c := &Collector{
		logger: config.Logger,

		awsClients:       awsClients,
		installationName: config.InstallationName,
	}

	return c, nil
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- vpcsDesc
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.logger.Log("level", "debug", "message", "collecting metrics")

	c.collectVPCs(ch)

	c.logger.Log("level", "debug", "message", "finished collecting metrics")
}
