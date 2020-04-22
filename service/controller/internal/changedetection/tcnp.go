package changedetection

import (
	"context"
	"fmt"
	"reflect"
	"sort"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

type TCNPConfig struct {
	Logger micrologger.Logger
}

// TCNP is a detection service implementation deciding if the TCNP stack should
// be updated.
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

// ShouldScale determines whether the reconciled TCNP stack should be scaled.
//
//     The node pool's scaling max changes.
//     The node pool's scaling min changes.
//
func (t *TCNP) ShouldScale(ctx context.Context, cr infrastructurev1alpha2.AWSMachineDeployment) (bool, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return false, microerror.Mask(err)
	}

	asgEmpty := cc.Status.TenantCluster.ASG.IsEmpty()
	asgMaxEqual := cc.Status.TenantCluster.ASG.MaxSize == key.MachineDeploymentScalingMax(cr)
	asgMinEqual := cc.Status.TenantCluster.ASG.MinSize == key.MachineDeploymentScalingMin(cr)

	if !asgEmpty && !asgMaxEqual {
		t.logger.LogCtx(ctx,
			"level", "debug",
			"message", "detected TCNP stack should scale up",
			"reason", fmt.Sprintf("scaling max changed from %d to %d", cc.Status.TenantCluster.ASG.MaxSize, key.MachineDeploymentScalingMax(cr)),
		)
		return true, nil
	}
	if !asgEmpty && !asgMinEqual {
		t.logger.LogCtx(ctx,
			"level", "debug",
			"message", "detected TCNP stack should scale down",
			"reason", fmt.Sprintf("scaling min changed from %d to %d", cc.Status.TenantCluster.ASG.MinSize, key.MachineDeploymentScalingMin(cr)),
		)
		return true, nil
	}

	return false, nil
}

// ShouldUpdate determines whether the reconciled TCNP stack should be updated.
//
//     The worker node's docker volume size changes.
//     The worker node's instance type changes.
//     The operator's version changes.
//     The composition of security groups changes.
//
func (t *TCNP) ShouldUpdate(ctx context.Context, cr infrastructurev1alpha2.AWSMachineDeployment) (bool, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return false, microerror.Mask(err)
	}

	dockerVolumeEqual := cc.Status.TenantCluster.TCNP.WorkerInstance.DockerVolumeSizeGB == key.MachineDeploymentDockerVolumeSizeGB(cr)
	instanceTypeEqual := cc.Status.TenantCluster.TCNP.WorkerInstance.Type == key.MachineDeploymentInstanceType(cr)
	operatorVersionEqual := cc.Status.TenantCluster.OperatorVersion == key.OperatorVersion(&cr)
	securityGroupsEqual := securityGroupsEqual(cc.Status.TenantCluster.TCNP.SecurityGroupIDs, cc.Spec.TenantCluster.TCNP.SecurityGroupIDs)

	if !dockerVolumeEqual {
		t.logger.LogCtx(ctx,
			"level", "debug",
			"message", "detected TCNP stack should update",
			"reason", fmt.Sprintf("worker instance docker volume size changed from %#q to %#q", cc.Status.TenantCluster.TCNP.WorkerInstance.DockerVolumeSizeGB, key.MachineDeploymentDockerVolumeSizeGB(cr)),
		)
		return true, nil
	}
	if !instanceTypeEqual {
		t.logger.LogCtx(ctx,
			"level", "debug",
			"message", "detected TCNP stack should update",
			"reason", fmt.Sprintf("worker instance type changed from %#q to %#q", cc.Status.TenantCluster.TCNP.WorkerInstance.Type, key.MachineDeploymentInstanceType(cr)),
		)
		return true, nil
	}
	if !operatorVersionEqual {
		t.logger.LogCtx(ctx,
			"level", "debug",
			"message", "detected TCNP stack should update",
			"reason", fmt.Sprintf("operator version changed from %#q to %#q", cc.Status.TenantCluster.OperatorVersion, key.OperatorVersion(&cr)),
		)
		return true, nil
	}
	if !securityGroupsEqual {
		t.logger.LogCtx(ctx,
			"level", "debug",
			"message", "detected TCNP stack should update",
			"reason", "security groups changed",
		)
		return true, nil
	}

	return false, nil
}

func securityGroupsEqual(cur []string, des []string) bool {
	sort.Strings(cur)
	sort.Strings(des)

	return reflect.DeepEqual(cur, des)
}
