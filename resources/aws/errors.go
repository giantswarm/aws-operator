package aws

import (
	"github.com/juju/errgo"
)

const (
	// Format for masked notFoundErrors.
	notFoundErrorFormat string = "%s with name %s not found"
	// Format for masked attributeEmptyError.
	attributeEmptyErrorFormat string = "attribute %s cannot be empty"
)

type resourceType string

const (
	ELBType           resourceType = "elb"
	HostedZoneType    resourceType = "hosted zone"
	GatewayType       resourceType = "gateway"
	InstanceType      resourceType = "instance"
	RouteTableType    resourceType = "route table"
	RouteType         resourceType = "route"
	SecurityGroupType resourceType = "security group"
	SubnetType        resourceType = "subnet"
	VPCType           resourceType = "vpc"
)

// NotFound errors.

var notFoundError = errgo.New("not found")

// IsNotFound asserts NotFoundError.
func IsNotFound(err error) bool {
	return errgo.Cause(err) == notFoundError
}

// Delete errors.

var resourceDeleteError = errgo.New("couldn't delete resource, it lacks the necessary data (ID)")

// IsResourceDelete asserts resourceDeleteError.
func IsResourceDelete(err error) bool {
	return errgo.Cause(err) == resourceDeleteError
}

// Other errors.

var clientNotInitializedError = errgo.New("the client has not been initialized")

// IsClientNotInitialized asserts clientNotInitializedError.
func IsClientNotInitialized(err error) bool {
	return errgo.Cause(err) == clientNotInitializedError
}

var keyPairCannotCreateAndNotFoundError = errgo.New("couldn't create and find the keypair")

// IsKeyPairCannotCreateAndNotFound asserts keyPairCannotCreateAndNotFoundError.
func IsKeyPairCannotCreateAndNotFound(err error) bool {
	return errgo.Cause(err) == keyPairCannotCreateAndNotFoundError
}

var notImplementedMethodError = errgo.New("not implemented method")

// IsNotImplementedMethod asserts notImplementedMethodError.
func IsNotImplementedMethod(err error) bool {
	return errgo.Cause(err) == notImplementedMethodError
}

var noBucketInBucketObjectError = errgo.New("object needs to belong to some bucket")

// IsNoBucketInBucketObject asserts noBucketInBucketObjectError.
func IsNoBucketInBucketObject(err error) bool {
	return errgo.Cause(err) == noBucketInBucketObjectError
}

var kmsKeyAliasEmptyError = errgo.New("the KMS key alias cannot be empty")

// IsKMSKeyAliasEmpty asserts kmsKeyAliasEmptyError.
func IsKMSKeyAliasEmpty(err error) bool {
	return errgo.Cause(err) == kmsKeyAliasEmptyError
}

var attributeEmptyError = errgo.New("attribute cannot be empty")

// IsPortsToOpenEmpty asserts portsToOpenEmptyError.
func IsAttributeEmpty(err error) bool {
	return errgo.Cause(err) == attributeEmptyError
}
