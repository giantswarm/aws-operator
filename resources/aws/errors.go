package aws

import (
	"github.com/juju/errgo"
)

var (
	notImplementedMethodError = errgo.New("not implemented")

	noBucketInBucketObjectError = errgo.New("Object needs to belong to some bucket")

	gatewayFindError       = errgo.New("Couldn't find gateway")
	routeTableFindError    = errgo.New("Couldn't find route table")
	securityGroupFindError = errgo.New("Couldn't find security group")
	subnetFindError        = errgo.New("Couldn't find subnet")
	vpcFindError           = errgo.New("Couldn't find VPC")

	resourceDeleteError = errgo.New("Couldn't delete resource, it lacks the necessary data (ID)")
)
