package cpvpc

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

const (
	Name = "cpvpc"
)

type Config struct {
	Logger micrologger.Logger

	InstallationName string
}

type Resource struct {
	logger micrologger.Logger

	cachedCidr string
	cachedID   string
	mutex      sync.Mutex

	installationName string
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationName must not be empty", config)
	}

	r := &Resource{
		logger: config.Logger,

		cachedCidr: "",
		cachedID:   "",
		mutex:      sync.Mutex{},

		installationName: config.InstallationName,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) addVPCInfoToContext(ctx context.Context) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		vpcCIDR, vpcID, err := r.lookup(ctx, cc.Client.ControlPlane.AWS.EC2, r.installationName)
		if err != nil {
			return microerror.Mask(err)
		}

		cc.Status.ControlPlane.VPC.CIDR = vpcCIDR
		cc.Status.ControlPlane.VPC.ID = vpcID
	}

	return nil
}

func (r *Resource) lookup(ctx context.Context, client EC2, installationName string) (string, string, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// We check if we have all VPC info cached for the requested installation.
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding cached vpc info for %#q", installationName))

		if r.cachedCidr != "" && r.cachedID != "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found cached vpc info %#q for %#q", r.cachedCidr, installationName))
			return r.cachedCidr, r.cachedID, nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find cached vpc info for %#q", installationName))
	}

	// We do not have a cached VPC Info for the requested installation. So we look it
	// up.
	var vpcCIDR string
	var vpcID string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding vpc info for %#q", installationName))

		i := &ec2.DescribeVpcsInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String(fmt.Sprintf("tag:%s", key.TagName)),
					Values: []*string{
						aws.String(installationName),
					},
				},
				{
					Name: aws.String(fmt.Sprintf("tag:%s", key.TagCluster)),
					Values: []*string{
						aws.String(installationName),
					},
				},
			},
		}

		o, err := client.DescribeVpcs(i)
		if err != nil {
			return "", "", microerror.Mask(err)
		}
		if len(o.Vpcs) != 1 {
			return "", "", microerror.Maskf(executionFailedError, "expected one vpc, got %d", len(o.Vpcs))
		}

		vpcCIDR = *o.Vpcs[0].CidrBlock
		vpcID = *o.Vpcs[0].VpcId

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found vpc cidr %#q and vpc id %#q for %#q", vpcCIDR, vpcID, installationName))
	}

	// At this point we found a VPC info and we cache it.
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("caching vpc info for %#q", installationName))
		r.cachedCidr = vpcCIDR
		r.cachedID = vpcID
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("cached vpc info for %#q", installationName))
	}

	return vpcCIDR, vpcID, nil
}
