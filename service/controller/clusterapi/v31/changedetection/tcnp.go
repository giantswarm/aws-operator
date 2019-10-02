package changedetection

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/key"
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

// ShouldScale determines whether the reconciled tenant cluster node pool should
// be scaled. A tenant cluster node pool is only allowed to scale in the
// following cases.
//
//     The node pool's scaling max changes.
//     The node pool's scaling min changes.
//
func (t *TCNP) ShouldScale(ctx context.Context, md v1alpha1.MachineDeployment) (bool, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return false, microerror.Mask(err)
	}

	asgEmpty := cc.Status.TenantCluster.TCNP.ASG.IsEmpty()
	asgMaxEqual := cc.Status.TenantCluster.TCNP.ASG.MaxSize == key.MachineDeploymentScalingMax(md)
	asgMinEqual := cc.Status.TenantCluster.TCNP.ASG.MinSize == key.MachineDeploymentScalingMin(md)

	if !asgEmpty && !asgMaxEqual {
		t.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("detected tenant cluster node pool should scale up due to scaling max changes from %d to %d", cc.Status.TenantCluster.TCNP.ASG.MaxSize, key.MachineDeploymentScalingMax(md)))
		return true, nil
	}
	if !asgEmpty && !asgMinEqual {
		t.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("detected tenant cluster node pool should scale down due to scaling min changes from %d to %d", cc.Status.TenantCluster.TCNP.ASG.MinSize, key.MachineDeploymentScalingMin(md)))
		return true, nil
	}

	return false, nil
}

// ShouldUpdate determines whether the reconciled tenant cluster node pool
// should be updated. A tenant cluster node pool is only allowed to update in
// the following cases.
//
//     The worker node's docker volume size changes.
//     The worker node's instance type changes.
//     The operator's version changes.
//
func (t *TCNP) ShouldUpdate(ctx context.Context, md v1alpha1.MachineDeployment) (bool, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return false, microerror.Mask(err)
	}

	dockerVolumeEqual := cc.Status.TenantCluster.TCNP.WorkerInstance.DockerVolumeSizeGB == key.MachineDeploymentDockerVolumeSizeGB(md)
	instanceTypeEqual := cc.Status.TenantCluster.TCNP.WorkerInstance.Type == key.MachineDeploymentInstanceType(md)
	operatorVersionEqual := cc.Status.TenantCluster.OperatorVersion == key.OperatorVersion(&md)

	if !dockerVolumeEqual {
		t.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("detected tenant cluster node pool should update due to worker instance docker volume size changes from %#q to %#q", cc.Status.TenantCluster.TCNP.WorkerInstance.DockerVolumeSizeGB, key.MachineDeploymentDockerVolumeSizeGB(md)))
		return true, nil
	}
	if !instanceTypeEqual {
		t.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("detected tenant cluster node pool should update due to worker instance type changes from %#q to %#q", cc.Status.TenantCluster.TCNP.WorkerInstance.Type, key.MachineDeploymentInstanceType(md)))
		return true, nil
	}
	if !operatorVersionEqual {
		t.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("detected tenant cluster node pool should update due to operator version changes from %#q to %#q", cc.Status.TenantCluster.OperatorVersion, key.OperatorVersion(&md)))
		return true, nil
	}

	return false, nil
}
