package changedetection

import (
	"context"
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/internal/recorder"
)

type TCCPConfig struct {
	Event  recorder.Interface
	Logger micrologger.Logger
}

// TCCP is a detection service implementation deciding if the TCCP stack should
// be updated.
type TCCP struct {
	event  recorder.Interface
	logger micrologger.Logger
}

func NewTCCP(config TCCPConfig) (*TCCP, error) {
	if config.Event == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Event must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	t := &TCCP{
		event:  config.Event,
		logger: config.Logger,
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

	if !azsEqual {
		t.logger.LogCtx(ctx,
			"level", "debug",
			"message", "detected TCCP stack should update",
			"reason", "availability zones changed",
		)
		t.event.Emit(ctx, &cr, "CFUpdateRequested", "detected TCCP stack should update: availability zones changed")
		return true, nil
	}
	if !operatorVersionEqual {
		t.logger.LogCtx(ctx,
			"level", "debug",
			"message", "detected TCCP stack should update",
			"reason", fmt.Sprintf("operator version changed from %#q to %#q", cc.Status.TenantCluster.OperatorVersion, key.OperatorVersion(&cr)),
		)
		t.event.Emit(ctx, &cr, "CFUpdateRequested", fmt.Sprintf("detected TCCP stack should update: operator version changed from %#q to %#q", cc.Status.TenantCluster.OperatorVersion, key.OperatorVersion(&cr)))
		return true, nil
	}

	return false, nil
}
