package detection

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/legacy/v27/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/legacy/v27/key"
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
func (d *Detection) ShouldScale(ctx context.Context, cr v1alpha1.AWSConfig) (bool, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return false, microerror.Mask(err)
	}

	if !cc.Status.TenantCluster.TCCP.ASG.IsEmpty() && cc.Status.TenantCluster.TCCP.ASG.MaxSize != key.ScalingMax(cr) {
		d.logger.LogCtx(ctx, "level", "debug", "message", "detected the tenant cluster should scale due to scaling max changes")
		return true, nil
	}
	if !cc.Status.TenantCluster.TCCP.ASG.IsEmpty() && cc.Status.TenantCluster.TCCP.ASG.MinSize != key.ScalingMin(cr) {
		d.logger.LogCtx(ctx, "level", "debug", "message", "detected the tenant cluster should scale due to scaling min changes")
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
func (d *Detection) ShouldUpdate(ctx context.Context, cr v1alpha1.AWSConfig) (bool, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return false, microerror.Mask(err)
	}

	if cc.Status.TenantCluster.MasterInstance.Type != key.MasterInstanceType(cr) {
		d.logger.LogCtx(ctx, "level", "debug", "message", "detected the tenant cluster should update due to master instance type changes")
		return true, nil
	}
	if cc.Status.TenantCluster.WorkerInstance.DockerVolumeSizeGB != key.WorkerDockerVolumeSizeGB(cr) {
		d.logger.LogCtx(ctx, "level", "debug", "message", "detected the tenant cluster should update due to worker instance docker volume size changes")
		return true, nil
	}
	if cc.Status.TenantCluster.WorkerInstance.Type != key.WorkerInstanceType(cr) {
		d.logger.LogCtx(ctx, "level", "debug", "message", "detected the tenant cluster should update due to worker instance type changes")
		return true, nil
	}
	if cc.Status.TenantCluster.VersionBundleVersion != key.VersionBundleVersion(cr) {
		d.logger.LogCtx(ctx, "level", "debug", "message", "detected the tenant cluster should update due to version bundle version changes")
		return true, nil
	}

	return false, nil
}
