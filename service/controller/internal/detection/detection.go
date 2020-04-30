package detection

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
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
		d.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("detected the tenant cluster should scale due to scaling max changes: cc.Status.TenantCluster.TCCP.ASG.MaxSize is %d while key.ScalingMax(cr) is %d", cc.Status.TenantCluster.TCCP.ASG.MaxSize, key.ScalingMax(cr)))
		return true, nil
	}
	if !cc.Status.TenantCluster.TCCP.ASG.IsEmpty() && cc.Status.TenantCluster.TCCP.ASG.MinSize != key.ScalingMin(cr) {
		d.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("detected the tenant cluster should scale due to scaling min changes: cc.Status.TenantCluster.TCCP.ASG.MinSize is %d while key.ScalingMin(cr) is %d", cc.Status.TenantCluster.TCCP.ASG.MinSize, key.ScalingMin(cr)))
		return true, nil
	}

	return false, nil
}

// ShouldUpdate determines whether the reconciled tenant cluster should be
// updated. A tenant cluster is only allowed to update in the following cases.
//
//   The master ignition hash changes.
//   The master instance AMI changes.
//   The master instance type changes.
//   The worker instance docker volume size changes.
//   The worker ignition hash changes.
//   The worker instance AMI changes.
//   The worker instance type changes.
//
func (d *Detection) ShouldUpdate(ctx context.Context, cr v1alpha1.AWSConfig) (bool, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return false, microerror.Mask(err)
	}

	imageID, err := key.ImageID(cr, cc.Spec.TenantCluster.Release)
	if err != nil {
		return false, microerror.Mask(err)
	}

	changeCandidates := []struct {
		name         string
		desiredValue string
		currentValue string
	}{
		{
			name:         "master ignition hash",
			desiredValue: cc.Spec.TenantCluster.MasterInstance.IgnitionHash,
			currentValue: cc.Status.TenantCluster.MasterInstance.IgnitionHash,
		},
		{
			name:         "master instance AMI",
			desiredValue: imageID,
			currentValue: cc.Status.TenantCluster.MasterInstance.Image,
		},
		{
			name:         "master instance type",
			desiredValue: key.MasterInstanceType(cr),
			currentValue: cc.Status.TenantCluster.MasterInstance.Type,
		},
		{
			name:         "worker instance docker volume size",
			desiredValue: key.WorkerDockerVolumeSizeGB(cr),
			currentValue: cc.Status.TenantCluster.WorkerInstance.DockerVolumeSizeGB,
		},
		{
			name:         "worker ignition hash",
			desiredValue: cc.Spec.TenantCluster.WorkerInstance.IgnitionHash,
			currentValue: cc.Status.TenantCluster.WorkerInstance.IgnitionHash,
		},
		{
			name:         "worker instance AMI",
			desiredValue: imageID,
			currentValue: cc.Status.TenantCluster.WorkerInstance.Image,
		},
		{
			name:         "worker instance type",
			desiredValue: key.WorkerInstanceType(cr),
			currentValue: cc.Status.TenantCluster.WorkerInstance.Type,
		},
	}

	template := "detected the tenant cluster should update due to changes in %s: current value of is %q while desired value is %q"
	for _, candidate := range changeCandidates {
		if candidate.desiredValue != candidate.currentValue {
			message := fmt.Sprintf(template, candidate.name, candidate.currentValue, candidate.desiredValue)
			d.logger.LogCtx(ctx, "level", "debug", "message", message)
			return true, nil
		}
	}

	return false, nil
}
