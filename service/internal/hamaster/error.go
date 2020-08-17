package hamaster

import "github.com/giantswarm/microerror"

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var notFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}

var tooManyCRsError = &microerror.Error{
	Kind: "tooManyCRsError",
	Desc: "There is only a single G8sControlPlane CR allowed with the current implementation.",
}

// IsTooManyCRsError asserts tooManyCRsError.
func IsTooManyCRsError(err error) bool {
	return microerror.Cause(err) == tooManyCRsError
}

var availabilityZonesNilError = &microerror.Error{
	Kind: "availabilityZonesNilError",
	Desc: "The availability zones in AWSControlPlane CR must not be nil.",
}

// IsAvalailabilityZonesNilError asserts tooManyCRsError.
func IsAvalailabilityZonesNilError(err error) bool {
	return microerror.Cause(err) == availabilityZonesNilError
}
