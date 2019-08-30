package changedetection

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
)

type TCNPConfig struct {
	Logger micrologger.Logger
}

// TCNP is a detection service implementation deciding if a node pool should be
// updated or scaled.
type TCNP struct {
	logger micrologger.Logger
}

func NewTCNP(config TCNPConfig) (*TCNP, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	t := &TCNP{
		logger: config.Logger,
	}

	return t, nil
}

// ShouldScale determines whether the reconciled node pool should be scaled. A
// node pool is only allowed to scale in the following cases.
//
//     The tenant cluster's scaling max changes.
//     The tenant cluster's scaling min changes.
//
func (t *TCNP) ShouldScale(ctx context.Context, md v1alpha1.MachineDeployment) (bool, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return false, microerror.Mask(err)
	}

	if !cc.Status.TenantCluster.TCNP.ASG.IsEmpty() && cc.Status.TenantCluster.TCNP.ASG.MaxSize != key.MachineDeploymentScalingMax(md) {
		t.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("detected node pool should scale up due to scaling max changes from %d to %d", cc.Status.TenantCluster.TCNP.ASG.MaxSize, key.MachineDeploymentScalingMax(md)))
		return true, nil
	}
	if !cc.Status.TenantCluster.TCNP.ASG.IsEmpty() && cc.Status.TenantCluster.TCNP.ASG.MinSize != key.MachineDeploymentScalingMin(md) {
		t.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("detected node pool should scale down due to scaling min changes from %d to %d", cc.Status.TenantCluster.TCNP.ASG.MinSize, key.MachineDeploymentScalingMin(md)))
		return true, nil
	}

	return false, nil
}

// ShouldUpdate determines whether the reconciled node pool should be updated. A
// node pool is only allowed to update in the following cases.
//
//     The worker node's docker volume size changes.
//     The worker node's instance type changes.
//     The tenant cluster's version changes.
//
func (t *TCNP) ShouldUpdate(ctx context.Context, md v1alpha1.MachineDeployment) (bool, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return false, microerror.Mask(err)
	}

	if cc.Status.TenantCluster.TCNP.WorkerInstance.DockerVolumeSizeGB != key.MachineDeploymentDockerVolumeSizeGB(md) {
		t.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("detected node pool should update due to worker instance docker volume size changes from %#q to %#q", cc.Status.TenantCluster.TCNP.WorkerInstance.DockerVolumeSizeGB, key.MachineDeploymentDockerVolumeSizeGB(md)))
		return true, nil
	}
	if cc.Status.TenantCluster.TCNP.WorkerInstance.Type != key.MachineDeploymentInstanceType(md) {
		t.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("detected node pool should update due to worker instance type changes from %#q to %#q", cc.Status.TenantCluster.TCNP.WorkerInstance.Type, key.MachineDeploymentInstanceType(md)))
		return true, nil
	}
	if cc.Status.TenantCluster.OperatorVersion != key.OperatorVersion(&md) {
		t.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("detected node pool should update due to operator version changes from %#q to %#q", cc.Status.TenantCluster.OperatorVersion, key.OperatorVersion(&md)))
		return true, nil
	}

	return false, nil
}
