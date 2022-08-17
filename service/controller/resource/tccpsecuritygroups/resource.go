package tccpsecuritygroups

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/v14/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v14/service/controller/key"
)

const (
	Name = "tccpsecuritygroups"
)

type Config struct {
	CtrlClient    client.Client
	Logger        micrologger.Logger
	ToClusterFunc func(ctx context.Context, v interface{}) (infrastructurev1alpha3.AWSCluster, error)
}

type Resource struct {
	ctrlClient    client.Client
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
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	r := &Resource{
		ctrlClient:    config.CtrlClient,
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

	cluster := apiv1beta1.Cluster{}
	err = r.ctrlClient.Get(ctx, client.ObjectKey{Namespace: cr.Namespace, Name: cr.Name}, &cluster)
	if err != nil {
		return microerror.Mask(err)
	}

	enableAWSCNI := key.IsAWSCNINeeded(cluster)

	var groups []*ec2.SecurityGroup
	{
		r.logger.Debugf(ctx, "finding security groups for tenant cluster %#q", key.ClusterID(&cr))

		filterValues := []*string{
			aws.String(key.SecurityGroupName(&cr, "internal-api")),
			aws.String(key.SecurityGroupName(&cr, "master")),
		}

		if enableAWSCNI {
			filterValues = append(filterValues, aws.String(key.SecurityGroupName(&cr, "aws-cni")))
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
		if enableAWSCNI {
			expectedLength = 3
		}

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
