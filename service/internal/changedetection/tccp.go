package changedetection

import (
	"context"
	"fmt"
	"reflect"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

type TCCPConfig struct {
	Logger micrologger.Logger
}

// TCCP is a detection service implementation deciding if the TCCP stack should
// be updated.
type TCCP struct {
	logger micrologger.Logger
}

func NewTCCP(config TCCPConfig) (*TCCP, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	t := &TCCP{
		logger: config.Logger,
	}

	return t, nil
}

// ShouldUpdate determines whether the reconciled TCCP stack should be updated.
//
//     The node pool's combined availability zone configuration changes.
//     The operator's version changes.
//
func (t *TCCP) ShouldUpdate(ctx context.Context, cr infrastructurev1alpha2.AWSCluster, tags map[string]string) (bool, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return false, microerror.Mask(err)
	}

	azsEqual := availabilityZonesEqual(cc.Spec.TenantCluster.TCCP.AvailabilityZones, cc.Status.TenantCluster.TCCP.AvailabilityZones)
	operatorVersionEqual := cc.Status.TenantCluster.OperatorVersion == key.OperatorVersion(&cr)
	tagsEqual := reflect.DeepEqual(cr.Labels, tags)

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
	if !tagsEqual {
		t.logger.LogCtx(ctx,
			"level", "debug",
			"message", "detected TCCP stack should update",
			"reason", fmt.Sprintf("tags have changed from %#q to %#q", cr.Labels, tags),
		)
		return true, nil
	}

	return false, nil
}
