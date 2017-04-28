package create

import (
	"github.com/juju/errgo"
)

var (
	invalidConfigError           = errgo.New("invalid config")
	secretsRetrievalFailedError  = errgo.New("could not retrieve secrets")
	missingCloudConfigKeyError   = errgo.New("missing required key in the cloudconfig")
	malformedCloudConfigKeyError = errgo.New("malformed key in the cloudconfig")
)

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return errgo.Cause(err) == invalidConfigError
}

func IsMissingCloudConfigKey(err error) bool {
	return errgo.Cause(err) == missingCloudConfigKeyError
}

func IsMalformedCloudConfigKey(err error) bool {
	return errgo.Cause(err) == malformedCloudConfigKeyError
}
