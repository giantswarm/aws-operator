package ipam

import (
	"context"
	"fmt"
	"net"

	"github.com/giantswarm/e2etests/ipam/provider"
	"github.com/giantswarm/ipam"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type Config struct {
	Logger   micrologger.Logger
	Provider provider.Interface

	ClusterID string
}

type IPAM struct {
	logger   micrologger.Logger
	provider provider.Interface

	clusterID string
}

func New(config Config) (*IPAM, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Provider == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}

	if config.ClusterID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterID must not be empty", config)
	}

	i := &IPAM{
		logger:   config.Logger,
		provider: config.Provider,

		clusterID: config.ClusterID,
	}

	return i, nil
}

func (i *IPAM) Test(ctx context.Context) error {
	// Clusters to be created during this test.
	var (
		clusterOne   = i.clusterID + "-1"
		clusterTwo   = i.clusterID + "-2"
		clusterThree = i.clusterID + "-3"
		clusterFour  = i.clusterID + "-4"
	)

	// Lists of clusters we work with in the test below.
	var (
		threeClusters = []string{clusterOne, clusterTwo, clusterThree}
		fourClusters  = []string{clusterOne, clusterTwo, clusterThree, clusterFour}
	)

	// Map of allocated subnets and tenant cluster ID pairs. The map keys are
	// subnets. The map values are cluster IDs.
	var (
		allocatedSubnets = map[string]string{}
	)

	defer func() {
		i.logger.LogCtx(ctx, "level", "debug", "message", "deleting all tenant clusters")

		for _, c := range fourClusters {
			err := i.provider.DeleteCluster(ctx, c)
			if err != nil {
				i.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("failed to delete tenant cluster %#q", c), "stack", fmt.Sprintf("%#v", microerror.Mask(err)))
			}
		}

		i.logger.LogCtx(ctx, "level", "debug", "message", "deleted all tenant clusters")
	}()

	{
		i.logger.LogCtx(ctx, "level", "debug", "message", "creating three tenant clusters")

		for _, c := range threeClusters {
			err := i.provider.CreateCluster(ctx, c)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		i.logger.LogCtx(ctx, "level", "debug", "message", "created three tenant clusters")
	}

	{
		i.logger.LogCtx(ctx, "level", "debug", "message", "waiting for three tenant clusters to be created")

		for _, c := range threeClusters {
			err := i.provider.WaitForClusterCreated(ctx, c)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		i.logger.LogCtx(ctx, "level", "debug", "message", "waited for three tenant clusters to be created")
	}

	{
		i.logger.LogCtx(ctx, "level", "debug", "message", "verifying subnet allocations do not overlap")

		for _, c := range threeClusters {
			s, err := i.provider.GetClusterStatus(ctx, c)
			if err != nil {
				return microerror.Mask(err)
			}

			otherCluster, exists := allocatedSubnets[s.Network.CIDR]
			if exists {
				return microerror.Maskf(alreadyExistsError, "subnet %s already exists for %s", s.Network.CIDR, otherCluster)
			}

			// Verify that allocated subnets don't overlap.
			for subnet, _ := range allocatedSubnets {
				err := verifyNoOverlap(s.Network.CIDR, subnet)
				if err != nil {
					return microerror.Mask(err)
				}
			}

			allocatedSubnets[s.Network.CIDR] = c
		}

		i.logger.LogCtx(ctx, "level", "debug", "message", "verified subnet allocations do not overlap")
	}

	{
		i.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting tenant cluster %#q", clusterTwo))

		err := i.provider.DeleteCluster(ctx, clusterTwo)
		if err != nil {
			return microerror.Mask(err)
		}

		i.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted tenant cluster %#q", clusterTwo))
	}

	{
		i.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating tenant cluster %#q", clusterFour))

		err := i.provider.CreateCluster(ctx, clusterFour)
		if err != nil {
			return microerror.Mask(err)
		}

		i.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created tenant cluster %#q", clusterFour))
	}

	{
		i.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waiting for tenant cluster %#q to be deleted", clusterTwo))

		err := i.provider.WaitForClusterDeleted(ctx, clusterTwo)
		if err != nil {
			return microerror.Mask(err)
		}

		i.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waited for tenant cluster %#q to be deleted", clusterTwo))
	}

	{
		i.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waiting for tenant cluster %#q to be created", clusterFour))

		err := i.provider.WaitForClusterCreated(ctx, clusterFour)
		if err != nil {
			return microerror.Mask(err)
		}

		i.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waited for tenant cluster %#q to be created", clusterFour))
	}

	{
		i.logger.LogCtx(ctx, "level", "debug", "message", "verifying subnet allocations do not overlap")

		s, err := i.provider.GetClusterStatus(ctx, clusterFour)
		if err != nil {
			return microerror.Mask(err)
		}

		otherCluster, exists := allocatedSubnets[s.Network.CIDR]
		if exists {
			return microerror.Maskf(alreadyExistsError, "subnet %s already exists for %s", s.Network.CIDR, otherCluster)
		}

		for subnet, _ := range allocatedSubnets {
			err := verifyNoOverlap(s.Network.CIDR, subnet)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		i.logger.LogCtx(ctx, "level", "debug", "message", "verified subnet allocations do not overlap")
	}

	return nil
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
