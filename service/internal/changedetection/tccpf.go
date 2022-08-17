package changedetection

import (
	"context"
	"fmt"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/v14/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v14/service/controller/key"
)

type TCCPFConfig struct {
	Logger micrologger.Logger
}

// TCCPF is a detection service implementation deciding if the TCCPF stack
// should be updated.
type TCCPF struct {
	logger micrologger.Logger
}

func NewTCCPF(config TCCPFConfig) (*TCCPF, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	t := &TCCPF{
		logger: config.Logger,
	}

	return t, nil
}

// ShouldUpdate determines whether the reconciled TCCPF stack should be updated.
//
//     The node pool's combined availability zone configuration changes.
//     The operator's version changes.
//
func (t *TCCPF) ShouldUpdate(ctx context.Context, cr infrastructurev1alpha3.AWSCluster) (bool, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return false, microerror.Mask(err)
	}

	azsEqual := availabilityZonesEqual(cc.Spec.TenantCluster.TCCP.AvailabilityZones, cc.Status.TenantCluster.TCCP.AvailabilityZones)
	operatorVersionEqual := cc.Status.TenantCluster.OperatorVersion == key.OperatorVersion(&cr)

	if !azsEqual {
		t.logger.LogCtx(ctx,
			"level", "debug",
			"message", "detected TCCPF stack should update",
			"reason", "availability zones changed",
		)
		return true, nil
	}
	if !operatorVersionEqual {
		t.logger.LogCtx(ctx,
			"level", "debug",
			"message", "detected TCCPF stack should update",
			"reason", fmt.Sprintf("operator version changed from %#q to %#q", cc.Status.TenantCluster.OperatorVersion, key.OperatorVersion(&cr)),
		)
		return true, nil
	}

	return false, nil
}
