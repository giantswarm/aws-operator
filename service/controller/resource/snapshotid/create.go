package snapshotid

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/meta"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
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

	i := &ec2.DescribeSnapshotsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String(fmt.Sprintf("tag:%s", key.TagCluster)),
				Values: []*string{
					aws.String(key.ClusterID(cr)),
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

	o, err := cc.Client.TenantCluster.AWS.EC2.DescribeSnapshots(i)
	if err != nil {
		return microerror.Mask(err)
	}
	count := len(o.Snapshots)
	if count == 0 {
		return microerror.Maskf(notExistsError, "snapshot")
	}
	if count != 1 {
		return microerror.Maskf(executionFailedError, "expected one snapshot, got %d", count)
	}

	cc.Status.TenantCluster.MasterInstance.EtcdVolumeSnapshotID = key.HAMasterSnapshotIDValue

	return nil
}
