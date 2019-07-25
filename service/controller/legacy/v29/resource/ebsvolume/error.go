package ebsvolume

import (
	"github.com/giantswarm/microerror"
)

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var volumeAttachedError = &microerror.Error{
	Kind: "volumeAttachedError",
}

// isVolumeAttached asserts volumeAttachedError.
func isVolumeAttached(err error) bool {
	return microerror.Cause(err) == volumeAttachedError
}
