package changedetection

import (
	"context"
	"fmt"
	"strings"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/internal/cloudtags"
	"github.com/giantswarm/aws-operator/service/internal/hamaster"
	"github.com/giantswarm/aws-operator/service/internal/recorder"
	"github.com/giantswarm/aws-operator/service/internal/releases"
)

type TCCPNConfig struct {
	CloudTags cloudtags.Interface
	Event     recorder.Interface
	HAMaster  hamaster.Interface
	Logger    micrologger.Logger
	Releases  releases.Interface
}

// TCCPN is a detection service implementation deciding if the TCCPN stack
// should be updated.
type TCCPN struct {
	cloudTags cloudtags.Interface
	haMaster  hamaster.Interface
	logger    micrologger.Logger
}

func NewTCCPN(config TCCPNConfig) (*TCCPN, error) {
	if config.CloudTags == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CloudTags must not be empty", config)
	}
	if config.Event == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Event must not be empty", config)
	}
	if config.HAMaster == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HAMaster must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Releases == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Releases must not be empty", config)
	}

	t := &TCCPN{
		cloudTags: config.CloudTags,
		event:     config.Event,
		haMaster:  config.HAMaster,
		logger:    config.Logger,
		releases:  config.Releases,
	}

	return t, nil
}

// ShouldUpdate determines whether the reconciled TCCPN stack should be updated.
//
//     The master node's instance type changes.
//     The operator's version changes.
//
func (t *TCCPN) ShouldUpdate(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane) (bool, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return false, microerror.Mask(err)
	}

	var rep int
	{
		rep, err = t.haMaster.Replicas(ctx, &cr)
		if err != nil {
			return false, microerror.Mask(err)
		}
	}

	var currentRelease releasev1alpha1.Release
	{
		currentRelease, err = t.releases.Release(ctx, cc.Status.TenantCluster.ReleaseVersion)
		if releases.IsNotFound(err) {
			// fall through
		} else if err != nil {
			return false, microerror.Mask(err)
		}
	}

	var targetRelease releasev1alpha1.Release
	{
		targetRelease, err = t.releases.Release(ctx, key.ReleaseVersion(&cr))
		if releases.IsNotFound(err) {
			// fall through
		} else if err != nil {
			return false, microerror.Mask(err)
		}
	}

	componentVersionsEqual := releaseComponentsEqual(currentRelease, targetRelease)
	masterInstanceEqual := cc.Status.TenantCluster.TCCPN.InstanceType == key.ControlPlaneInstanceType(cr)
	masterReplicasEqual := cc.Status.TenantCluster.TCCPN.MasterReplicas == rep
	operatorVersionEqual := cc.Status.TenantCluster.OperatorVersion == key.OperatorVersion(&cr)

	cloudTagsNotEqual, err := t.cloudTags.CloudTagsNotInSync(ctx, &cr.ObjectMeta, key.StackTCCPN)
	if err != nil {
		return false, microerror.Mask(err)
	}

	if !componentVersionsEqual {
		t.logger.LogCtx(ctx,
			"level", "debug",
			"message", "detected TCCPN stack should update",
			"reason", strings.Join(componentsDiff(currentRelease, targetRelease), ", "),
		)
		t.event.Emit(ctx, &cr, "CFUpdateRequested", fmt.Sprintf("detected TCCPN stack should update: %s", strings.Join(componentsDiff(currentRelease, targetRelease), ", ")))
		return true, nil
	}
	if !masterInstanceEqual {
		t.logger.LogCtx(
			ctx,
			"level", "debug",
			"message", "detected TCCPN stack should update",
			"reason", fmt.Sprintf("master instance type changed from %#q to %#q", cc.Status.TenantCluster.TCCPN.InstanceType, key.ControlPlaneInstanceType(cr)),
		)
		t.event.Emit(ctx, &cr, "CFUpdateRequested", fmt.Sprintf("detected TCCPN stack should update: master instance type changed from %#q to %#q", cc.Status.TenantCluster.TCCPN.InstanceType, key.ControlPlaneInstanceType(cr)))
		return true, nil
	}
	if !masterReplicasEqual {
		t.logger.LogCtx(
			ctx,
			"level", "debug",
			"message", "detected TCCPN stack should update",
			"reason", fmt.Sprintf("master replicas changed from %d to %d", cc.Status.TenantCluster.TCCPN.MasterReplicas, rep),
		)
		t.event.Emit(ctx, &cr, "CFUpdateRequested", fmt.Sprintf("detected TCCPN stack should update: master replicas changed from %d to %d", cc.Status.TenantCluster.TCCPN.MasterReplicas, rep))
		return true, nil
	}
	if !operatorVersionEqual {
		t.logger.LogCtx(ctx,
			"level", "debug",
			"message", "detected TCCPN stack should update",
			"reason", fmt.Sprintf("operator version changed from %#q to %#q", cc.Status.TenantCluster.OperatorVersion, key.OperatorVersion(&cr)),
		)
		t.event.Emit(ctx, &cr, "CFUpdateRequested", fmt.Sprintf("detected TCCPN stack should update: operator version changed from %#q to %#q", cc.Status.TenantCluster.OperatorVersion, key.OperatorVersion(&cr)))
		return true, nil
	}
	if cloudTagsNotEqual {
		t.logger.LogCtx(ctx,
			"level", "debug",
			"message", "detected TCCPN stack should update",
			"reason", "cloud tags are not synced",
		)
		return true, nil
	}

	return false, nil
}
