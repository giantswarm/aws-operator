package collector

import (
	"context"
	"sync"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/kubernetes"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
)

const (
	Namespace = "aws_operator"
)

type Config struct {
	G8sClient versioned.Interface
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	AwsConfig             awsutil.Config
	InstallationName      string
	TrustedAdvisorEnabled bool
}

type Collector struct {
	g8sClient versioned.Interface
	k8sClient kubernetes.Interface
	logger    micrologger.Logger

	bootOnce sync.Once

	awsConfig             awsutil.Config
	installationName      string
	trustedAdvisorEnabled bool
}

func New(config Config) (*Collector, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
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

	c := &Collector{
		g8sClient: config.G8sClient,
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		bootOnce: sync.Once{},

		awsConfig:             config.AwsConfig,
		installationName:      config.InstallationName,
		trustedAdvisorEnabled: config.TrustedAdvisorEnabled,
	}

	return c, nil
}

func (c *Collector) Boot(ctx context.Context) {
	c.bootOnce.Do(func() {
		if c.trustedAdvisorEnabled {
			prometheus.MustRegister(trustedAdvisorError)

			prometheus.MustRegister(getChecksDuration)
			prometheus.MustRegister(getResourcesDuration)

			c.logger.Log("level", "debug", "message", "trusted advisor metrics collection enabled")
		} else {
			c.logger.Log("level", "debug", "message", "trusted advisor metrics collection disabled")
		}
	})
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- clustersDesc

	ch <- serviceLimit
	ch <- serviceUsage

	if c.trustedAdvisorEnabled {
		ch <- trustedAdvisorSupport
	}

	ch <- vpcsDesc
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.logger.Log("level", "debug", "message", "collecting metrics")

	// Get aws clients
	clients, err := c.getAWSClients()
	if err != nil {
		c.logger.Log("level", "error", "message", "could not get aws clients", "error", err.Error())
	}

	var wg sync.WaitGroup

	collectFuncs := []func(chan<- prometheus.Metric, []awsutil.Clients){
		c.collectClusterInfo,
		c.collectAccountsVPCs,
	}

	if c.trustedAdvisorEnabled {
		collectFuncs = append(collectFuncs, c.collectAccountsTrustedAdvisorChecks)
	}

	for _, collectFunc := range collectFuncs {
		wg.Add(1)

		go func(collectFunc func(chan<- prometheus.Metric, []awsutil.Clients)) {
			defer wg.Done()
			collectFunc(ch, clients)
		}(collectFunc)
	}

	wg.Wait()

	c.logger.Log("level", "debug", "message", "finished collecting metrics")
}
