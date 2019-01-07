package scaling

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/e2etests/scaling/provider"
)

type Config struct {
	Logger   micrologger.Logger
	Provider provider.Interface
}

type Scaling struct {
	logger   micrologger.Logger
	provider provider.Interface
}

func New(config Config) (*Scaling, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Provider == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}

	s := &Scaling{
		logger:   config.Logger,
		provider: config.Provider,
	}

	return s, nil
}

func (s *Scaling) Test(ctx context.Context) error {
	var err error

	var numMasters int
	{
		s.logger.LogCtx(ctx, "level", "debug", "message", "looking for the number of masters")

		numMasters, err = s.provider.NumMasters()
		if err != nil {
			return microerror.Mask(err)
		}

		s.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d masters", numMasters))
	}

	var numWorkers int
	{
		s.logger.LogCtx(ctx, "level", "debug", "message", "looking for the number of workers")

		numWorkers, err = s.provider.NumWorkers()
		if err != nil {
			return microerror.Mask(err)
		}

		s.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d workers", numWorkers))
	}

	{
		s.logger.LogCtx(ctx, "level", "debug", "message", "scaling up one worker")

		err = s.provider.AddWorker()
		if err != nil {
			return microerror.Mask(err)
		}

		s.logger.LogCtx(ctx, "level", "debug", "message", "scaled up one worker")
	}

	{
		s.logger.LogCtx(ctx, "level", "debug", "message", "waiting for scaling up to be complete")

		err = s.provider.WaitForNodes(ctx, numMasters+numWorkers+1)
		if err != nil {
			return microerror.Mask(err)
		}

		s.logger.LogCtx(ctx, "level", "debug", "message", "scaling up complete")
	}

	{
		s.logger.LogCtx(ctx, "level", "debug", "message", "scaling down one worker")

		err = s.provider.RemoveWorker()
		if err != nil {
			return microerror.Mask(err)
		}

		s.logger.LogCtx(ctx, "level", "debug", "message", "scaled down one worker")
	}

	{
		s.logger.LogCtx(ctx, "level", "debug", "message", "waiting for scaling down to be complete")

		err = s.provider.WaitForNodes(ctx, numMasters+numWorkers)
		if err != nil {
			return microerror.Mask(err)
		}

		s.logger.LogCtx(ctx, "level", "debug", "message", "scaling down complete")
	}

	return nil
}
