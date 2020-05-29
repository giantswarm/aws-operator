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

func availabilityZonesEqual(spec []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone, status []controllercontext.ContextStatusTenantClusterTCCPAvailabilityZone) bool {
	if spec == nil && status != nil {
		return false
	}

	if spec != nil && status == nil {
		return false
	}

	if len(spec) != len(status) {
		return false
	}

	for _, az := range spec {
		// alternatively could sort the slice and compare as before.
		if !containsAZ(az, status) {
			return false
		}
	}

	return true
}

// true if status slice has an AZ that is equal to target.
func containsAZ(target controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone, status []controllercontext.ContextStatusTenantClusterTCCPAvailabilityZone) bool {
	for _, az := range status {
		if availabilityZoneEqual(target, az) {
			return true
		}
	}
	return false
}

func availabilityZoneEqual(spec controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone, status controllercontext.ContextStatusTenantClusterTCCPAvailabilityZone) bool {
	if spec.Name != status.Name {
		return false
	}

	if spec.Subnet.Private.CIDR.String() != status.Subnet.Private.CIDR.String() {
		return false
	}

	if spec.Subnet.Private.ID != status.Subnet.Private.ID {
		return false
	}

	if spec.Subnet.Public.CIDR.String() != status.Subnet.Public.CIDR.String() {
		return false
	}

	if spec.Subnet.Public.ID != status.Subnet.Public.ID {
		return false
	}

	return true
}
