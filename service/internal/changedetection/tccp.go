package changedetection

import (
	"context"
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/internal/releases"
)

type TCCPConfig struct {
	Logger   micrologger.Logger
	Releases releases.Interface
}

// TCCP is a detection service implementation deciding if the TCCP stack should
// be updated.
type TCCP struct {
	logger   micrologger.Logger
	releases releases.Interface
}

func NewTCCP(config TCCPConfig) (*TCCP, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Releases == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Release must not be empty", config)
	}

	t := &TCCP{
		logger:   config.Logger,
		releases: config.Releases,
	}

	return t, nil
}

// ShouldUpdate determines whether the reconciled TCCP stack should be updated.
//
//     The node pool's combined availability zone configuration changes.
//     The operator's version changes.
//
func (t *TCCP) ShouldUpdate(ctx context.Context, cr infrastructurev1alpha2.AWSCluster) (bool, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return false, microerror.Mask(err)
	}

	azsEqual := availabilityZonesEqual(cc.Spec.TenantCluster.TCCP.AvailabilityZones, cc.Status.TenantCluster.TCCP.AvailabilityZones)
	operatorVersionEqual := cc.Status.TenantCluster.OperatorVersion == key.OperatorVersion(&cr)
	currentRelease, err := t.releases.Release(ctx, cc.Status.TenantCluster.ReleaseVersion)
	if err != nil {
		return false, microerror.Mask(err)
	}
	targetRelease, err := t.releases.Release(ctx, key.ReleaseVersion(&cr))
	if err != nil {
		return false, microerror.Mask(err)
	}
	_ = releaseComponentsEqual(currentRelease, targetRelease)

	if !azsEqual {
		t.logger.LogCtx(ctx,
			"level", "debug",
			"message", "detected TCCP stack should update",
			"reason", "availability zones changed",
		)
		return true, nil
	}
	if !operatorVersionEqual {
		t.logger.LogCtx(ctx,
			"level", "debug",
			"message", "detected TCCP stack should update",
			"reason", fmt.Sprintf("operator version changed from %#q to %#q", cc.Status.TenantCluster.OperatorVersion, key.OperatorVersion(&cr)),
		)
		return true, nil
	}

	return false, nil
}
