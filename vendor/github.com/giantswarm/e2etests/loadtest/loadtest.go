package loadtest

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/e2etests/loadtest/provider"
)

type Config struct {
	Logger   micrologger.Logger
	Provider provider.Interface
}

type LoadTest struct {
	logger   micrologger.Logger
	provider provider.Interface
}

func New(config Config) (*LoadTest, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Provider == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}

	s := &LoadTest{
		logger:   config.Logger,
		provider: config.Provider,
	}

	return s, nil
}

func (l *LoadTest) Test(ctx context.Context) error {
	var err error

	{
		l.logger.LogCtx(ctx, "level", "debug", "message", "installing loadtest-app")

		err = l.provider.InstallTestApp(ctx)
		if err != nil {
			return microerror.Mask(err)
		}

		l.logger.LogCtx(ctx, "level", "debug", "message", "installed loadtest-app")
	}

	{
		l.logger.LogCtx(ctx, "level", "debug", "message", "waiting for loadtest-app to be ready")

		err = l.provider.WaitForTestApp(ctx)
		if err != nil {
			return microerror.Mask(err)
		}

		l.logger.LogCtx(ctx, "level", "debug", "message", "loadtest-app is ready")
	}

	return nil
}
