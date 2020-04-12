package changedetection

import (
	"context"
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

type TCCPNConfig struct {
	Logger micrologger.Logger
}

// TCCPN is a detection service implementation deciding if the TCCPN stack
// should be updated.
type TCCPN struct {
	logger micrologger.Logger
}

func NewTCCPN(config TCCPNConfig) (*TCCPN, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	t := &TCCPN{
		logger: config.Logger,
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

	masterInstanceEqual := cc.Status.TenantCluster.TCCPN.InstanceType == key.ControlPlaneInstanceType(cr)
	operatorVersionEqual := cc.Status.TenantCluster.OperatorVersion == key.OperatorVersion(&cr)

	if !masterInstanceEqual {
		t.logger.LogCtx(
			ctx,
			"level", "debug",
			"message", "detected TCCPN stack should update",
			"reason", fmt.Sprintf("master instance type changed from %#q to %#q", cc.Status.TenantCluster.TCCPN.InstanceType, key.ControlPlaneInstanceType(cr)),
		)
		return true, nil
	}
	if !operatorVersionEqual {
		t.logger.LogCtx(ctx,
			"level", "debug",
			"message", "detected TCCPN stack should update",
			"reason", fmt.Sprintf("operator version changed from %#q to %#q", cc.Status.TenantCluster.OperatorVersion, key.OperatorVersion(&cr)),
		)
		return true, nil
	}

	return false, nil
}
