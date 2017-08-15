package aws

import (
	"github.com/giantswarm/microerror"
)

const (
	// Format for masked notFoundErrors.
	notFoundErrorFormat string = "%s with name %s not found"
	// Format for masked attributeEmptyError.
	attributeEmptyErrorFormat string = "attribute %s cannot be empty"
)

type resourceType string

const (
	CloudFormationStackType resourceType = "cloud formation stack"
	ELBType                 resourceType = "elb"
	HostedZoneType          resourceType = "hosted zone"
	GatewayType             resourceType = "gateway"
	InstanceType            resourceType = "instance"
	LaunchConfigurationType resourceType = "launch configuration"
	RouteTableType          resourceType = "route table"
	RouteType               resourceType = "route"
	SecurityGroupType       resourceType = "security group"
	SubnetType              resourceType = "subnet"
	VPCType                 resourceType = "vpc"
)

// NotFound errors.

var notFoundError = microerror.New("not found")

// IsNotFound asserts NotFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}

var tooManyResultsError = microerror.New("too many results")

// IsTooManyResults asserts tooManyResultsError.
func IsTooManyResults(err error) bool {
	return microerror.Cause(err) == tooManyResultsError
}

// Delete errors.

var resourceDeleteError = microerror.New("couldn't delete resource, it lacks the necessary data (ID)")

// IsResourceDelete asserts resourceDeleteError.
func IsResourceDelete(err error) bool {
	return microerror.Cause(err) == resourceDeleteError
}

// Other errors.

var clientNotInitializedError = microerror.New("the client has not been initialized")

// IsClientNotInitialized asserts clientNotInitializedError.
func IsClientNotInitialized(err error) bool {
	return microerror.Cause(err) == clientNotInitializedError
}

var keyPairCannotCreateAndNotFoundError = microerror.New("couldn't create and find the keypair")

// IsKeyPairCannotCreateAndNotFound asserts keyPairCannotCreateAndNotFoundError.
func IsKeyPairCannotCreateAndNotFound(err error) bool {
	return microerror.Cause(err) == keyPairCannotCreateAndNotFoundError
}

var notImplementedMethodError = microerror.New("not implemented method")

// IsNotImplementedMethod asserts notImplementedMethodError.
func IsNotImplementedMethod(err error) bool {
	return microerror.Cause(err) == notImplementedMethodError
}

var noBucketInBucketObjectError = microerror.New("object needs to belong to some bucket")

// IsNoBucketInBucketObject asserts noBucketInBucketObjectError.
func IsNoBucketInBucketObject(err error) bool {
	return microerror.Cause(err) == noBucketInBucketObjectError
}

var kmsKeyAliasEmptyError = microerror.New("the KMS key alias cannot be empty")

// IsKMSKeyAliasEmpty asserts kmsKeyAliasEmptyError.
func IsKMSKeyAliasEmpty(err error) bool {
	return microerror.Cause(err) == kmsKeyAliasEmptyError
}

var attributeEmptyError = microerror.New("attribute cannot be empty")

// IsPortsToOpenEmpty asserts portsToOpenEmptyError.
func IsAttributeEmpty(err error) bool {
	return microerror.Cause(err) == attributeEmptyError
}
