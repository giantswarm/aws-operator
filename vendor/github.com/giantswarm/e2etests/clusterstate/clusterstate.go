package clusterstate

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/e2etests/clusterstate/provider"
)

type Config struct {
	Logger   micrologger.Logger
	Provider provider.Interface
}

type ClusterState struct {
	logger   micrologger.Logger
	provider provider.Interface
}

func New(config Config) (*ClusterState, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Provider == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}

	s := &ClusterState{
		logger:   config.Logger,
		provider: config.Provider,
	}

	return s, nil
}

func (c *ClusterState) Test(ctx context.Context) error {
	var err error

	{
		c.logger.LogCtx(ctx, "level", "debug", "message", "installing e2e-app")

		err = c.provider.InstallTestApp()
		if err != nil {
			return microerror.Mask(err)
		}

		c.logger.LogCtx(ctx, "level", "debug", "message", "installed e2e-app")
	}

	{
		c.logger.LogCtx(ctx, "level", "debug", "message", "waiting for guest cluster")

		err = c.provider.WaitForGuestReady()
		if err != nil {
			return microerror.Mask(err)
		}

		c.logger.LogCtx(ctx, "level", "debug", "message", "guest cluster ready")
	}

	{
		c.logger.LogCtx(ctx, "level", "debug", "message", "rebooting master")

		err = c.provider.RebootMaster()
		if err != nil {
			return microerror.Mask(err)
		}

		c.logger.LogCtx(ctx, "level", "debug", "message", "rebooted master")
	}

	{
		c.logger.LogCtx(ctx, "level", "debug", "message", "waiting for guest cluster")

		err = c.provider.WaitForGuestReady()
		if err != nil {
			return microerror.Mask(err)
		}

		c.logger.LogCtx(ctx, "level", "debug", "message", "guest cluster ready")
	}

	{
		c.logger.LogCtx(ctx, "level", "debug", "message", "replacing master node")

		err = c.provider.ReplaceMaster()
		if err != nil {
			return microerror.Mask(err)
		}

		c.logger.LogCtx(ctx, "level", "debug", "message", "master node replaced")
	}

	{
		c.logger.LogCtx(ctx, "level", "debug", "message", "waiting for guest cluster")

		err = c.provider.WaitForGuestReady()
		if err != nil {
			return microerror.Mask(err)
		}

		c.logger.LogCtx(ctx, "level", "debug", "message", "guest cluster ready")
	}

	return nil
}
