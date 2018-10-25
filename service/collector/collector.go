package collector

import (
	"context"
	"fmt"
	"sync"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"

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
	Helper *Helper
	Logger micrologger.Logger

	InstallationName      string
	TrustedAdvisorEnabled bool
}

type Collector struct {
	helper *Helper
	logger micrologger.Logger

	bootOnce sync.Once

	installationName      string
	trustedAdvisorEnabled bool
}

func New(config Config) (*Collector, error) {
	if config.Helper == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Helper must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationName must not be empty", config)
	}

	c := &Collector{
		helper: config.Helper,
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
	clients, err := c.helper.GetAWSClients()
	if err != nil {
		c.logger.Log("level", "error", "message", "could not get aws clients", "error", err.Error())
	}

	var wg sync.WaitGroup

	collectFuncs := []func(chan<- prometheus.Metric, []clientaws.Clients){
		c.collectClusterInfo,
		c.collectAccountsVPCs,
	}

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
