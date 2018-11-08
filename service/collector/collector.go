package collector

import (
	"context"
	"fmt"
	"sync"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/kubernetes"

	clientaws "github.com/giantswarm/aws-operator/client/aws"
)

const (
	GaugeValue float64 = 1
	Namespace          = "aws_operator"
)

const (
	ClusterTag      = "giantswarm.io/cluster"
	InstallationTag = "giantswarm.io/installation"
	OrganizationTag = "giantswarm.io/organization"
)

const (
	ClusterLabel      = "cluster_id"
	InstallationLabel = "installation"
	OrganizationLabel = "organization"
)

// NOTE the collector implementation below is deprecated. Further collector
// implementations should align with the exporterkit interface and be configured
// in the collector set list. See also service/service.go.

type Config struct {
	G8sClient versioned.Interface
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	AWSConfig             clientaws.Config
	InstallationName      string
	TrustedAdvisorEnabled bool
}

type Collector struct {
	helper *helper
	logger micrologger.Logger

	bootOnce sync.Once

	installationName      string
	trustedAdvisorEnabled bool
}

func New(config Config) (*Collector, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationName must not be empty", config)
	}

	var err error

	var h *helper
	{
		c := helperConfig{
			G8sClient: config.G8sClient,
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			AWSConfig: config.AWSConfig,
		}

		h, err = newHelper(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &Collector{
		helper: h,
		logger: config.Logger,

		bootOnce: sync.Once{},

		installationName:      config.InstallationName,
		trustedAdvisorEnabled: config.TrustedAdvisorEnabled,
	}

	return c, nil
}

func (c *Collector) Boot(ctx context.Context) {
	c.bootOnce.Do(func() {
		{
			c.logger.LogCtx(ctx, "level", "debug", "message", "registering collector")

			err := prometheus.Register(c)
			if IsAlreadyRegisteredError(err) {
				c.logger.LogCtx(ctx, "level", "debug", "message", "collector already registered")
			} else if err != nil {
				c.logger.Log("level", "error", "message", "registering collector failed", "stack", fmt.Sprintf("%#v", err))
			} else {
				c.logger.LogCtx(ctx, "level", "debug", "message", "registered collector")
			}
		}

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
	ch <- serviceLimit
	ch <- serviceUsage

	if c.trustedAdvisorEnabled {
		ch <- trustedAdvisorSupport
	}
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.logger.Log("level", "debug", "message", "collecting metrics")

	// Get aws clients
	clients, err := c.helper.GetAWSClients()
	if err != nil {
		c.logger.Log("level", "error", "message", "could not get aws clients", "error", err.Error())
	}

	var wg sync.WaitGroup

	collectFuncs := []func(chan<- prometheus.Metric, []clientaws.Clients){}

	if c.trustedAdvisorEnabled {
		collectFuncs = append(collectFuncs, c.collectAccountsTrustedAdvisorChecks)
	}

	for _, collectFunc := range collectFuncs {
		wg.Add(1)

		go func(collectFunc func(chan<- prometheus.Metric, []clientaws.Clients)) {
			defer wg.Done()
			collectFunc(ch, clients)
		}(collectFunc)
	}

	wg.Wait()

	c.logger.Log("level", "debug", "message", "finished collecting metrics")
}
