package vpccidr

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/controllercontext"
)

const (
	Name = "vpccidrv27"
)

type Config struct {
	Logger micrologger.Logger

	VPCPeerID string
}

type Resource struct {
	logger micrologger.Logger

	// cidrs is a mapping of vpcs IDs and CIDRs, where the key is the ID
	// and the value is the CIDR.
	cidrs map[string]string
	mutex sync.Mutex

	vpcPeerID string
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.VPCPeerID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.VPCPeerID must not be empty", config)
	}

	r := &Resource{
		logger: config.Logger,

		cidrs: map[string]string{},
		mutex: sync.Mutex{},

		vpcPeerID: config.VPCPeerID,
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
		cidr, err := r.lookup(ctx, cc.Client.ControlPlane.AWS.EC2, r.vpcPeerID)
		if err != nil {
			return microerror.Mask(err)
		}

		cc.Status.ControlPlane.VPC.CIDR = cidr
	}

	return nil
}

func (r *Resource) lookup(ctx context.Context, client EC2, vpc string) (string, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// We check if we have a VPC CIDR cached for the requested VPC ID. If we find
	// one, we return the result.
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding cached vpc cidr for %#q", vpc))

		cidr, ok := r.cidrs[vpc]
		if ok {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found cached vpc cidr %#q for %#q", cidr, vpc))
			return cidr, nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find cached vpc cidr for %#q", vpc))
	}

	// We do not have a cached VPC CIDR for the requested VPC ID. So we look it
	// up.
	var cidr string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding vpc cidr for %#q", vpc))

		i := &ec2.DescribeVpcsInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("vpc-id"),
					Values: []*string{
						aws.String(vpc),
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

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found vpc cidr %#q for %#q", cidr, vpc))
	}

	// At this point we found a VPC CIDR and can cache it using the requested VPC
	// ID.
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("caching vpc cidr %#q for %#q", cidr, vpc))
		r.cidrs[vpc] = cidr
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("cached vpc cidr %#q for %#q", cidr, vpc))
	}

	return cidr, nil
}
