package collector

import (
	"sync"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
)

const (
	Namespace = "aws_operator"
)

type Config struct {
	G8sClient versioned.Interface
	Logger    micrologger.Logger

	AwsConfig        awsutil.Config
	InstallationName string
}

type Collector struct {
	g8sClient versioned.Interface
	logger    micrologger.Logger

	awsClients       awsutil.Clients
	installationName string
}

func New(config Config) (*Collector, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
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
		g8sClient: config.G8sClient,
		logger:    config.Logger,

		awsClients:       awsClients,
		installationName: config.InstallationName,
	}

	return c, nil
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- clustersDesc

	ch <- serviceLimit
	ch <- serviceUsage
	ch <- trustedAdvisorSupport

	ch <- vpcsDesc
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.logger.Log("level", "debug", "message", "collecting metrics")

	collectFuncs := []func(chan<- prometheus.Metric){
		c.collectClusterInfo,
		c.collectTrustedAdvisorChecks,
		c.collectVPCs,
	}

	var wg sync.WaitGroup

	for _, collectFunc := range collectFuncs {
		wg.Add(1)

		go func(collectFunc func(ch chan<- prometheus.Metric)) {
			defer wg.Done()
			collectFunc(ch)
		}(collectFunc)
	}

	wg.Wait()

	c.logger.Log("level", "debug", "message", "finished collecting metrics")
}
