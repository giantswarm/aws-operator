package workerasgname

import "github.com/giantswarm/microerror"

var invalidConfigError = microerror.New("invalid config")

// IsInsserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}
