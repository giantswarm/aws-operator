package memory

import (
	"github.com/giantswarm/microerror"
)

var notFoundError = microerror.New("not found")

func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}
