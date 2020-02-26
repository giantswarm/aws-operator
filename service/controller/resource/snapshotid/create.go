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

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding the tenant cluster snapshot id")

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
		r.logger.LogCtx(ctx, "level", "debug", "message", "the tenant cluster snapshot id is not available yet")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	}
	if count > 1 {
		return microerror.Maskf(executionFailedError, "expected one snapshot, got %d", count)
	}

	//store the snapshot id
	snapshot := o.Snapshots[0]
	snapshotID := aws.StringValue(snapshot.SnapshotId)
	cc.Status.TenantCluster.MasterInstance.EtcdVolumeSnapshotID = snapshotID
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found the tenant cluster snapshot id %#q", snapshotID))

	return nil
}
