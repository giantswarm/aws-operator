package snapshotid

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/api/pkg/key"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/meta"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	clusterID := key.ClusterID(cr)

	// TODO define key function

	// TODO lookup snapshot
	input := &ec2.DescribeSnapshotsInput{
		//use the filter here
		Filters: []*ec2.Filter{
			{
				Name: aws.String(fmt.Sprintf("tag:%s", key.TagCluster)),
				Values: []*string{
					aws.String(key.ClusterID(&cr)),
				},
			},
			{
				Name: aws.String(fmt.Sprintf("tag:%s", key.TagSnapshot)),
				Values: []*string{
					aws.String(key.HAMasterSnapshotIDValue),
				},
			},
		},
	}
	result, err := cc.Client.TenantCluster.AWS.EC2.DescribeSnapshots(input)
	// 1 is good
	// 0 bad
	// > 1 bad
	if err != nil {
		return microerror.Mask(err)
	}

	// TODO define structure
	cc.Status.TenantCluster.TCNP.ASG.MinSize = minSize

	return nil
}
