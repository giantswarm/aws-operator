package ipam

import (
	"context"
	"fmt"
	"net"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/e2etests/ipam/provider"
	"github.com/giantswarm/ipam"
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
		// Clusters to be created during this test.
		clusterOne   = c.hostClusterName + "-cluster0"
		clusterTwo   = c.hostClusterName + "-cluster1"
		clusterThree = c.hostClusterName + "-cluster2"
		clusterFour  = c.hostClusterName + "-cluster3"

		// allocatedSubnets[clusterName]subnetCIDRStr
		allocatedSubnets = make(map[string]string)
		err              error
		// guestFrameworks[clusterName]guestFramework
		guestFrameworks = make(map[string]*framework.Guest)
	)

	defer func() {
		c.logger.LogCtx(ctx, "level", "debug", "message", "ensuring all guest clusters possibly created in test are deleted.")

		for _, cn := range []string{clusterOne, clusterTwo, clusterThree, clusterFour} {
			err := c.provider.DeleteCluster(cn)
			if err != nil {
				c.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("cluster %s deletion failed: %#v", cn, err))
			}
		}
	}()

	// clusters to create in first batch
	clusters := []string{clusterOne, clusterTwo, clusterThree}

	{

		c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating three guest clusters: %#v", clusters))

		for _, cn := range clusters {
			err = c.provider.CreateCluster(cn)
			if err != nil {
				return microerror.Mask(err)
			}
			c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("create guest cluster %s", cn))
		}
	}

	{
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
		}

		c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("all three guest clusters are ready: %#v", clusters))
	}

	{
		c.logger.LogCtx(ctx, "level", "debug", "message", "verifying that subnet allocations don't overlap")

		for _, cn := range clusters {
			awsConfig, err := c.hostFramework.AWSCluster(cn)
			if err != nil {
				return microerror.Mask(err)
			}

			c.logger.LogCtx(ctx, "level", "debug", "message", "verify that there are no duplicate subnet allocations")
			subnet := awsConfig.Status.Cluster.Network.CIDR
			otherCluster, exists := allocatedSubnets[subnet]
			if exists {
				return microerror.Maskf(alreadyExistsError, "subnet %s already exists for %s", subnet, otherCluster)
			}

			// Verify that allocated subnets don't overlap.
			for k, _ := range allocatedSubnets {
				err := verifyNoOverlap(subnet, k)
				if err != nil {
					return microerror.Mask(err)
				}
			}

			allocatedSubnets[subnet] = cn
		}
	}

	{
		c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("terminating guest cluster %s and immediately creating new guest cluster %s", clusterTwo, clusterFour))

		err = c.provider.DeleteCluster(clusterTwo)
		if err != nil {
			return microerror.Mask(err)
		}

		err = c.provider.CreateCluster(clusterFour)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waiting for guest cluster %s to shutdown", clusterTwo))

		guest := guestFrameworks[clusterTwo]
		err = guest.WaitForAPIDown()
		if err != nil {
			return microerror.Mask(err)
		}
		delete(guestFrameworks, clusterTwo)

		c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("guest cluster %s down", clusterTwo))
	}

	{
		c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waiting for guest cluster %s to become up", clusterFour))

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
		c.logger.LogCtx(ctx, "level", "debug", "message", "verify that clusterFour subnet doesn't overlap with other allocations")
		awsConfig, err := c.hostFramework.AWSCluster(clusterFour)
		if err != nil {
			return microerror.Mask(err)
		}

		subnet := awsConfig.Status.Cluster.Network.CIDR
		otherCluster, exists := allocatedSubnets[subnet]
		if exists {
			return microerror.Maskf(alreadyExistsError, "subnet %s already exists for %s", subnet, otherCluster)
		}

		for k, _ := range allocatedSubnets {
			err := verifyNoOverlap(subnet, k)
			if err != nil {
				return microerror.Mask(err)
			}
		}
	}

	return err
}

func verifyNoOverlap(subnet1, subnet2 string) error {
	_, net1, err := net.ParseCIDR(subnet1)
	if err != nil {
		return err
	}

	_, net2, err := net.ParseCIDR(subnet2)
	if err != nil {
		return err
	}

	if ipam.Contains(*net1, *net2) {
		return microerror.Maskf(subnetsOverlapError, "subnet %s contains %s", net1, net2)
	}

	if ipam.Contains(*net2, *net1) {
		return microerror.Maskf(subnetsOverlapError, "subnet %s contains %s", net2, net1)
	}

	return nil
}
