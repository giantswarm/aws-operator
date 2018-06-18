package masternode

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/e2etests/masternode/provider"
)

type Config struct {
	Logger   micrologger.Logger
	Provider provider.Interface
}

type MasterNode struct {
	logger   micrologger.Logger
	provider provider.Interface
}

func New(config Config) (*MasterNode, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Provider == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}

	s := &MasterNode{
		logger:   config.Logger,
		provider: config.Provider,
	}

	return s, nil
}

func (m *MasterNode) Test(ctx context.Context) error {
	var err error

	{
		m.logger.LogCtx(ctx, "level", "debug", "message", "installing e2e-app")

		err = m.provider.InstallTestApp()
		if err != nil {
			return microerror.Mask(err)
		}

		m.logger.LogCtx(ctx, "level", "debug", "message", "installed e2e-app")
	}

	{
		m.logger.LogCtx(ctx, "level", "debug", "message", "waiting for guest cluster")

		err = m.provider.WaitForGuestReady()
		if err != nil {
			return microerror.Mask(err)
		}

		m.logger.LogCtx(ctx, "level", "debug", "message", "guest cluster ready")
	}

	{
		m.logger.LogCtx(ctx, "level", "debug", "message", "rebooting master")

		err = m.provider.RebootMaster()
		if err != nil {
			return microerror.Mask(err)
		}

		m.logger.LogCtx(ctx, "level", "debug", "message", "rebooted master")
	}

	{
		m.logger.LogCtx(ctx, "level", "debug", "message", "waiting for guest cluster")

		err = m.provider.WaitForGuestReady()
		if err != nil {
			return microerror.Mask(err)
		}

		m.logger.LogCtx(ctx, "level", "debug", "message", "guest cluster ready")
	}

	{
		m.logger.LogCtx(ctx, "level", "debug", "message", "replacing master node")

		err = m.provider.ReplaceMaster()
		if err != nil {
			return microerror.Mask(err)
		}

		m.logger.LogCtx(ctx, "level", "debug", "message", "master node replaced")
	}

	{
		m.logger.LogCtx(ctx, "level", "debug", "message", "waiting for guest cluster")

		err = m.provider.WaitForGuestReady()
		if err != nil {
			return microerror.Mask(err)
		}

		m.logger.LogCtx(ctx, "level", "debug", "message", "guest cluster ready")
	}

	return nil
}
