package detection

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/key"
)

type Config struct {
	Logger micrologger.Logger
}

// Detection is a service implementation deciding if a tenant cluster should be
// updated or scaled.
type Detection struct {
	logger micrologger.Logger
}

func New(config Config) (*Detection, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	d := &Detection{
		logger: config.Logger,
	}

	return d, nil
}

// ShouldScale determines whether the reconciled tenant cluster should be
// scaled. A tenant cluster is only allowed to scale in the following cases.
//
//     The tenant cluster's scaling max changes.
//     The tenant cluster's scaling min changes.
//
func (d *Detection) ShouldScale(ctx context.Context, md v1alpha1.MachineDeployment) (bool, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return false, microerror.Mask(err)
	}

	if !cc.Status.TenantCluster.TCCP.ASG.IsEmpty() && cc.Status.TenantCluster.TCCP.ASG.MaxSize != key.WorkerScalingMax(md) {
		d.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("detected the tenant cluster should scale due to scaling max changes: cc.Status.TenantCluster.TCCP.ASG.MaxSize is %d while key.WorkerScalingMax(md) is %d", cc.Status.TenantCluster.TCCP.ASG.MaxSize, key.WorkerScalingMax(md)))
		return true, nil
	}
	if !cc.Status.TenantCluster.TCCP.ASG.IsEmpty() && cc.Status.TenantCluster.TCCP.ASG.MinSize != key.WorkerScalingMin(md) {
		d.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("detected the tenant cluster should scale due to scaling min changes: cc.Status.TenantCluster.TCCP.ASG.MinSize is %d while key.WorkerScalingMin(md) is %d", cc.Status.TenantCluster.TCCP.ASG.MinSize, key.WorkerScalingMin(md)))
		return true, nil
	}

	return false, nil
}

// ShouldUpdate determines whether the reconciled tenant cluster should be
// updated. A tenant cluster is only allowed to update in the following cases.
//
//     The master node's instance type changes.
//     The worker node's docker volume size changes.
//     The worker node's instance type changes.
//     The tenant cluster's version changes.
//
func (d *Detection) ShouldUpdate(ctx context.Context, cl v1alpha1.Cluster, md v1alpha1.MachineDeployment) (bool, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return false, microerror.Mask(err)
	}

	if cc.Status.TenantCluster.MasterInstance.Type != key.MasterInstanceType(cl) {
		d.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("detected the tenant cluster should update due to master instance type changes: cc.Status.TenantCluster.MasterInstance.Type is %q while key.MasterInstanceType(cl) is %q", cc.Status.TenantCluster.MasterInstance.Type, key.MasterInstanceType(cl)))
		return true, nil
	}
	if cc.Status.TenantCluster.WorkerInstance.DockerVolumeSizeGB != key.WorkerDockerVolumeSizeGB(md) {
		d.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("detected the tenant cluster should update due to worker instance docker volume size changes: cc.Status.TenantCluster.WorkerInstance.DockerVolumeSizeGB is %q while key.WorkerDockerVolumeSizeGB(md) is %q", cc.Status.TenantCluster.WorkerInstance.DockerVolumeSizeGB, key.WorkerDockerVolumeSizeGB(md)))
		return true, nil
	}
	if cc.Status.TenantCluster.WorkerInstance.Type != key.WorkerInstanceType(md) {
		d.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("detected the tenant cluster should update due to worker instance type changes: cc.Status.TenantCluster.WorkerInstance.Type is %q while key.WorkerInstanceType(md) is %q", cc.Status.TenantCluster.WorkerInstance.Type, key.WorkerInstanceType(md)))
		return true, nil
	}
	if cc.Status.TenantCluster.VersionBundleVersion != key.ClusterVersion(cl) {
		d.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("detected the tenant cluster should update due to version bundle version changes: cc.Status.TenantCluster.VersionBundleVersion is %q while key.ClusterVersion(md) is %q", cc.Status.TenantCluster.VersionBundleVersion, key.ClusterVersion(cl)))
		return true, nil
	}

	return false, nil
}
