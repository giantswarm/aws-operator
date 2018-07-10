package ipam

import (
	"context"

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

	// First create all three guest clusters
	for _, cn := range clusters {
		err = c.provider.RequestGuestClusterCreation(cn)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	guestFrameworks := make(map[string]*framework.Guest)

	// Wait for them to be ready.
	for _, clusterName := range clusters {
		cfg := framework.GuestConfig{
			ClusterName: clusterName,
			Logger:      c.logger,
		}
		guestFramework, err := framework.NewGuest(cfg)
		if err != nil {
			return microerror.Mask(err)
		}

		guestFrameworks[clusterName] = guestFramework
		err = guestFramework.Setup()
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// TODO: Verify subnet properties for three created clusters.

	// Now terminate guest cluster #2 and create guest cluster #4.
	c.provider.RequestGuestClusterDeletion(clusterTwo)
	err = c.provider.RequestGuestClusterCreation(clusterFour)
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
	}

	// TODO: Verify that fourth cluster subnet allocation doesn't overlap with
	// terminated second cluster.

	for _, cn := range []string{clusterOne, clusterThree, clusterFour} {
		c.provider.RequestGuestClusterDeletion(cn)
		delete(guestFrameworks, cn)
	}

	return err
}
