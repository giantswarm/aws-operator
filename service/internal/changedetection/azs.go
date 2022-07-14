package changedetection

import (
	"github.com/giantswarm/aws-operator/v12/service/controller/controllercontext"
)

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
