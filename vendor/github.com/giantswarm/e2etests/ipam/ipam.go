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
	Logger   micrologger.Logger
	Provider provider.Interface
}

type IPAM struct {
	logger   micrologger.Logger
	provider provider.Interface
}

func New(config Config) (*IPAM, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Provider == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}

	s := &IPAM{
		logger:   config.Logger,
		provider: config.Provider,
	}

	return s, nil
}

func (c *IPAM) Test(ctx context.Context) error {
	var err error

	const (
		clusterOne   = "cluster0"
		clusterTwo   = "cluster1"
		clusterThree = "cluster2"
		clusterFour  = "cluster3"
	)

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
			ClusterName: cn,
			Logger:      c.logger,
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

	// TODO: Verify subnet properties for three created clusters.

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
			ClusterName: clusterFour,
			Logger:      c.logger,
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

	// TODO: Verify that fourth cluster subnet allocation doesn't overlap with
	// terminated second cluster.

	clusters = []string{clusterOne, clusterThree, clusterFour}

	c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting remaining guest clusters: %#v", clusters))

	for _, cn := range clusters {
		c.provider.DeleteCluster(cn)
		delete(guestFrameworks, cn)
	}

	return err
}
