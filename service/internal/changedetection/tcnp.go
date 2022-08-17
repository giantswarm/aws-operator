package changedetection

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	releasev1alpha1 "github.com/giantswarm/release-operator/v3/api/v1alpha1"

	"github.com/giantswarm/aws-operator/v13/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v13/service/controller/key"
	"github.com/giantswarm/aws-operator/v13/service/internal/recorder"
	"github.com/giantswarm/aws-operator/v13/service/internal/releases"
)

type TCNPConfig struct {
	Event    recorder.Interface
	Logger   micrologger.Logger
	Releases releases.Interface
}

// TCNP is a detection service implementation deciding if the TCNP stack should
// be updated.
type TCNP struct {
	event    recorder.Interface
	logger   micrologger.Logger
	releases releases.Interface
}

func NewTCNP(config TCNPConfig) (*TCNP, error) {
	if config.Event == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Event must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Releases == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Releases must not be empty", config)
	}

	t := &TCNP{
		event:    config.Event,
		logger:   config.Logger,
		releases: config.Releases,
	}

	return t, nil
}

// ShouldScale determines whether the reconciled TCNP stack should be scaled.
//
//     The node pool's scaling max changes.
//     The node pool's scaling min changes.
//
func (t *TCNP) ShouldScale(ctx context.Context, cr infrastructurev1alpha3.AWSMachineDeployment) (bool, error) {
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
func (t *TCNP) ShouldUpdate(ctx context.Context, cr infrastructurev1alpha3.AWSMachineDeployment) (bool, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return false, microerror.Mask(err)
	}

	var currentRelease releasev1alpha1.Release
	{
		currentRelease, err = t.releases.Release(ctx, cc.Status.TenantCluster.ReleaseVersion)
		if err != nil {
			return false, microerror.Mask(err)
		}
	}

	var targetRelease releasev1alpha1.Release
	{
		targetRelease, err = t.releases.Release(ctx, key.ReleaseVersion(&cr))
		if err != nil {
			return false, microerror.Mask(err)
		}
	}

	componentVersionsEqual := releaseComponentsEqual(currentRelease, targetRelease)
	dockerVolumeEqual := cc.Status.TenantCluster.TCNP.WorkerInstance.DockerVolumeSizeGB == key.MachineDeploymentDockerVolumeSizeGB(cr)
	instanceTypeEqual := cc.Status.TenantCluster.TCNP.WorkerInstance.Type == key.MachineDeploymentInstanceType(cr)
	operatorVersionEqual := cc.Status.TenantCluster.OperatorVersion == key.OperatorVersion(&cr)
	securityGroupsEqual := securityGroupsEqual(cc.Status.TenantCluster.TCNP.SecurityGroupIDs, cc.Spec.TenantCluster.TCNP.SecurityGroupIDs)

	if !componentVersionsEqual {
		t.logger.LogCtx(ctx,
			"level", "debug",
			"message", "detected TCNP stack should update",
			"reason", strings.Join(componentsDiff(currentRelease, targetRelease), ", "),
		)
		t.event.Emit(ctx, &cr, "CFUpdateRequested", fmt.Sprintf("detected TCNP stack should update: %s", strings.Join(componentsDiff(currentRelease, targetRelease), ", ")))
		return true, nil
	}
	if !dockerVolumeEqual {
		t.logger.LogCtx(ctx,
			"level", "debug",
			"message", "detected TCNP stack should update",
			"reason", fmt.Sprintf("worker instance docker volume size changed from %#q to %#q", cc.Status.TenantCluster.TCNP.WorkerInstance.DockerVolumeSizeGB, key.MachineDeploymentDockerVolumeSizeGB(cr)),
		)
		t.event.Emit(ctx, &cr, "CFUpdateRequested", fmt.Sprintf("detected TCNP stack should update: worker instance docker volume size changed from %#q to %#q", cc.Status.TenantCluster.TCNP.WorkerInstance.DockerVolumeSizeGB, key.MachineDeploymentDockerVolumeSizeGB(cr)))
		return true, nil
	}
	if !instanceTypeEqual {
		t.logger.LogCtx(ctx,
			"level", "debug",
			"message", "detected TCNP stack should update",
			"reason", fmt.Sprintf("worker instance type changed from %#q to %#q", cc.Status.TenantCluster.TCNP.WorkerInstance.Type, key.MachineDeploymentInstanceType(cr)),
		)
		t.event.Emit(ctx, &cr, "CFUpdateRequested", fmt.Sprintf("detected TCNP stack should update: worker instance type changed from %#q to %#q", cc.Status.TenantCluster.TCNP.WorkerInstance.Type, key.MachineDeploymentInstanceType(cr)))
		return true, nil
	}
	if !operatorVersionEqual {
		t.logger.LogCtx(ctx,
			"level", "debug",
			"message", "detected TCNP stack should update",
			"reason", fmt.Sprintf("operator version changed from %#q to %#q", cc.Status.TenantCluster.OperatorVersion, key.OperatorVersion(&cr)),
		)
		t.event.Emit(ctx, &cr, "CFUpdateRequested", fmt.Sprintf("detected TCNP stack should update: operator version changed from %#q to %#q", cc.Status.TenantCluster.OperatorVersion, key.OperatorVersion(&cr)))
		return true, nil
	}
	if !securityGroupsEqual {
		t.logger.LogCtx(ctx,
			"level", "debug",
			"message", "detected TCNP stack should update",
			"reason", "security groups changed",
		)
		t.event.Emit(ctx, &cr, "CFUpdateRequested", "detected TCNP stack should update: security groups changed")
		return true, nil
	}

	return false, nil
}

func securityGroupsEqual(cur []string, des []string) bool {
	sort.Strings(cur)
	sort.Strings(des)

	return reflect.DeepEqual(cur, des)
}
