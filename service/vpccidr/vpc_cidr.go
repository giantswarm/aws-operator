package vpccidr

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type Config struct {
	EC2    EC2
	Logger micrologger.Logger
}

type VPCCIDR struct {
	ec2    EC2
	logger micrologger.Logger

	// cidrs is a mapping of vpcs IDs and CIDRs, where the key is the ID
	// and the value is the CIDR.
	cidrs map[string]string
	mutex sync.Mutex
}

func New(config Config) (*VPCCIDR, error) {
	if config.EC2 == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.EC2 must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	c := &VPCCIDR{
		ec2:    config.EC2,
		logger: config.Logger,

		cidrs: map[string]string{},
		mutex: sync.Mutex{},
	}

	return c, nil
}

func (c *VPCCIDR) Lookup(ctx context.Context, vpc string) (string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	cidr, ok := c.cidrs[vpc]
	if ok {
		return cidr, nil
	}

	cidr, err := c.lookup(ctx, vpc)
	if err != nil {
		return "", microerror.Mask(err)
	}
	c.cidrs[vpc] = cidr

	return cidr, nil
}

func (c *VPCCIDR) lookup(ctx context.Context, vpc string) (string, error) {
	c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding vpc cidr for %#q", vpc))

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

	o, err := c.ec2.DescribeVpcs(i)
	if err != nil {
		return "", microerror.Mask(err)
	}
	if len(o.Vpcs) != 1 {
		return "", microerror.Maskf(executionFailedError, "expected one vpc, got %d", len(o.Vpcs))
	}

	cidr := *o.Vpcs[0].CidrBlock

	c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found vpc cidr %#q for %#q", cidr, vpc))

	return cidr, nil
}
