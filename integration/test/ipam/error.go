package ipam

import (
	"github.com/giantswarm/microerror"
)

var clusterCRStillExistsError = &microerror.Error{
	Desc: "The tenant cluster's CR still exists which indicates the cluster deletion is not complete.",
	Kind: "clusterCRStillExistsError",
}

// IsClusterCRStillExists asserts clusterCRStillExistsError.
func IsClusterCRStillExists(err error) bool {
	return microerror.Cause(err) == clusterCRStillExistsError
}

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var missingCreatedConditionError = &microerror.Error{
	Desc: "The tenant cluster's CR does not obtain the Created status condition to indicated the successfull cluster creation.",
	Kind: "missingCreatedConditionError",
}

// IsMissingCreatedCondition asserts missingCreatedConditionError.
func IsMissingCreatedCondition(err error) bool {
	return microerror.Cause(err) == missingCreatedConditionError
}
