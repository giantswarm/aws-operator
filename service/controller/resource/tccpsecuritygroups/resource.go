package tccpsecuritygroups

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/v13/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v13/service/controller/key"
)

const (
	Name = "tccpsecuritygroups"
)

type Config struct {
	Logger        micrologger.Logger
	ToClusterFunc func(ctx context.Context, v interface{}) (infrastructurev1alpha3.AWSCluster, error)
}

type Resource struct {
	logger        micrologger.Logger
	toClusterFunc func(ctx context.Context, v interface{}) (infrastructurev1alpha3.AWSCluster, error)
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.ToClusterFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterFunc must not be empty", config)
	}
	r := &Resource{
		logger:        config.Logger,
		toClusterFunc: config.ToClusterFunc,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) addInfoToCtx(ctx context.Context, cr infrastructurev1alpha3.AWSCluster) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var groups []*ec2.SecurityGroup
	{
		r.logger.Debugf(ctx, "finding security groups for tenant cluster %#q", key.ClusterID(&cr))

		filterValues := []*string{
			aws.String(key.SecurityGroupName(&cr, "internal-api")),
			aws.String(key.SecurityGroupName(&cr, "master")),
		}

		i := &ec2.DescribeSecurityGroupsInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("tag:Name"),
					Values: filterValues,
				},
				{
					Name: aws.String(fmt.Sprintf("tag:%s", key.TagStack)),
					Values: []*string{
						aws.String(key.StackTCCP),
					},
				},
			},
		}

		o, err := cc.Client.TenantCluster.AWS.EC2.DescribeSecurityGroups(i)
		if err != nil {
			return microerror.Mask(err)
		}

		groups = o.SecurityGroups

		expectedLength := 2

		if len(groups) > expectedLength {
			return microerror.Maskf(executionFailedError, "expected %d security groups, got %d", expectedLength, len(groups))
		}

		if len(groups) < expectedLength {
			r.logger.Debugf(ctx, "found %d out of expected %d security groups for tenant cluster %#q yet", len(groups), expectedLength, key.ClusterID(&cr))
			r.logger.Debugf(ctx, "canceling resource")

			return nil
		}

		r.logger.Debugf(ctx, "found security groups for tenant cluster %#q", key.ClusterID(&cr))
	}

	{
		cc.Status.TenantCluster.TCCP.SecurityGroups = groups
	}

	return nil
}
