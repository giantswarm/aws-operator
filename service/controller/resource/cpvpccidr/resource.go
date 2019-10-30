package cpvpccidr

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
)

const (
	Name = "cpvpccidr"
)

type Config struct {
	Logger micrologger.Logger

	InstallationName string
}

type Resource struct {
	logger micrologger.Logger

	cachedCidr string
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
		mutex:      sync.Mutex{},

		installationName: config.InstallationName,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) addVPCCIDRToContext(ctx context.Context) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		cidr, err := r.lookup(ctx, cc.Client.ControlPlane.AWS.EC2, r.installationName)
		if err != nil {
			return microerror.Mask(err)
		}

		cc.Status.ControlPlane.VPC.CIDR = cidr
	}

	return nil
}

func (r *Resource) lookup(ctx context.Context, client EC2, installationName string) (string, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// We check if we have a VPC CIDR cached for the requested installation. If we find
	// one, we return the result.
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding cached vpc cidr for %#q", installationName))

		if r.cachedCidr != "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found cached vpc cidr %#q for %#q", r.cachedCidr, installationName))
			return r.cachedCidr, nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find cached vpc cidr for %#q", installationName))
	}

	// We do not have a cached VPC CIDR for the requested installation. So we look it
	// up.
	var cidr string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding vpc cidr for %#q", installationName))

		i := &ec2.DescribeVpcsInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("giantswarm.io/installation"),
					Values: []*string{
						aws.String(installationName),
					},
				},
			},
		}

		o, err := client.DescribeVpcs(i)
		if err != nil {
			return "", microerror.Mask(err)
		}
		if len(o.Vpcs) != 1 {
			return "", microerror.Maskf(executionFailedError, "expected one vpc, got %d", len(o.Vpcs))
		}

		cidr = *o.Vpcs[0].CidrBlock

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found vpc cidr %#q for %#q", cidr, installationName))
	}

	// At this point we found a VPC CIDR and can cache it using the requested VPC
	// ID.
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("caching vpc cidr %#q for %#q", cidr, installationName))
		r.cachedCidr = cidr
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("cached vpc cidr %#q for %#q", cidr, installationName))
	}

	return cidr, nil
}
