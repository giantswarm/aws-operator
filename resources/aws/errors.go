package aws

import (
	"github.com/juju/errgo"
)

var (
	notImplementedMethodError = errgo.New("not implemented")

	noBucketInBucketObjectError = errgo.New("Object needs to belong to some bucket")

	securityGroupCreateAndFindError = errgo.New("Couldn't create security group, but couldn't find the existing one either")
)
