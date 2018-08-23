package ipam

import (
	"context"
	"fmt"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/e2etests/ipam/provider"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type Config struct {
	HostFramework *framework.Host
	Logger        micrologger.Logger
	Provider      provider.Interface

	CommonDomain    string
	HostClusterName string
}

type IPAM struct {
	hostFramework *framework.Host
	logger        micrologger.Logger
	provider      provider.Interface

	commonDomain    string
	hostClusterName string
}

func New(config Config) (*IPAM, error) {
	if config.HostFramework == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HostFramework must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Provider == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}
	if config.CommonDomain == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.CommonDomain must not be empty", config)
	}
	if config.HostClusterName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.HostClusterName must not be empty", config)
	}

	s := &IPAM{
		hostFramework: config.HostFramework,
		logger:        config.Logger,
		provider:      config.Provider,

		commonDomain:    config.CommonDomain,
		hostClusterName: config.HostClusterName,
	}

	return s, nil
}

func (c *IPAM) Test(ctx context.Context) error {

	var (
		clusterOne   = c.hostClusterName + "-cluster0"
		clusterTwo   = c.hostClusterName + "-cluster1"
		clusterThree = c.hostClusterName + "-cluster2"
		clusterFour  = c.hostClusterName + "-cluster3"

		// allocatedSubnets[clusterName]subnetCIDRStr
		allocatedSubnets = make(map[string]string)
		err              error
	)

	defer func() {
		c.logger.LogCtx(ctx, "level", "debug", "message", "ensuring all guest clusters possibly created in test are deleted.")

		for _, cn := range []string{clusterOne, clusterTwo, clusterThree, clusterFour} {
			c.provider.DeleteCluster(cn)
		}
	}()

	clusters := []string{clusterOne, clusterTwo, clusterThree}

	c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating three guest clusters: %#v", clusters))

	for _, cn := range clusters {
		err = c.provider.CreateCluster(cn)
		if err != nil {
			return microerror.Mask(err)
		}
		c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("create guest cluster %s", cn))
	}

	guestFrameworks := make(map[string]*framework.Guest)

	c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waiting for three guest clusters to be ready: %#v", clusters))

	for _, cn := range clusters {
		cfg := framework.GuestConfig{
			Logger: c.logger,

			ClusterID:    cn,
			CommonDomain: c.commonDomain,
		}
		guestFramework, err := framework.NewGuest(cfg)
		if err != nil {
			return microerror.Mask(err)
		}

		guestFrameworks[cn] = guestFramework
		err = guestFramework.Setup()
		if err != nil {
			return microerror.Mask(err)
		}
		c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("guest cluster %s ready", cn))

		awsConfig, err := c.hostFramework.AWSCluster(cn)
		if err != nil {
			return microerror.Mask(err)
		}

		// Verify that there are no duplicate subnet allocations.
		subnet := awsConfig.Status.Cluster.Network.CIDR
		otherCluster, exists := allocatedSubnets[subnet]
		if exists {
			return microerror.Maskf(alreadyExistsError, "subnet %s already exists for %s", subnet, otherCluster)
		}
		allocatedSubnets[subnet] = cn
	}

	c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("terminating guest cluster %s and immediately creating new guest cluster %s", clusterTwo, clusterFour))

	c.provider.DeleteCluster(clusterTwo)
	err = c.provider.CreateCluster(clusterFour)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		guest := guestFrameworks[clusterTwo]
		err = guest.WaitForAPIDown()
		if err != nil {
			return microerror.Mask(err)
		}
		delete(guestFrameworks, clusterTwo)
		c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("guest cluster %s down", clusterTwo))
	}

	{
		cfg := framework.GuestConfig{
			Logger: c.logger,

			ClusterID:    clusterFour,
			CommonDomain: c.commonDomain,
		}
		guestFramework, err := framework.NewGuest(cfg)
		if err != nil {
			return microerror.Mask(err)
		}

		guestFrameworks[clusterFour] = guestFramework
		err = guestFramework.Setup()
		if err != nil {
			return microerror.Mask(err)
		}
		c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("guest cluster %s up", clusterFour))
	}

	{
		// Verify that allocated subnet for clusterFour doesn't overlap with
		// terminated clusterTwo or any other existing cluster.
		awsConfig, err := c.hostFramework.AWSCluster(clusterFour)
		if err != nil {
			return microerror.Mask(err)
		}

		subnet := awsConfig.Status.Cluster.Network.CIDR
		otherCluster, exists := allocatedSubnets[subnet]
		if exists {
			return microerror.Maskf(alreadyExistsError, "subnet %s already exists for %s", subnet, otherCluster)
		}
	}

	return err
}
