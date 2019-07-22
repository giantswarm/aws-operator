package vpcid

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := r.toClusterFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	clusterID := key.ClusterID(&cr)

	i := &ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String(key.TagCluster),
				Values: []*string{
					aws.String(clusterID),
				},
			},
		},
	}
	o, err := cc.Client.TenantCluster.AWS.EC2.DescribeVpcs(i)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(o.Vpcs) > 1 {
		return microerror.Mask(tooManyResultsError)
	}

	if len(o.Vpcs) == 1 {
		if o.Vpcs[0].VpcId != nil {
			cc.Status.TenantCluster.TCCP.VPC.ID = *o.Vpcs[0].VpcId
		}
	}

	return nil
}
