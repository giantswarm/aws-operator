package aws

import (
	"fmt"

	"github.com/juju/errgo"
)

var (
	notImplementedMethodError = errgo.New("not implemented")

	noBucketInBucketObjectError = errgo.New("Object needs to belong to some bucket")

	resourceDeleteError       = errgo.New("Couldn't delete resource, it lacks the necessary data (ID)")
	clientNotInitializedError = errgo.New("The client has not been initialized")
)

type DomainNamedResourceNotFoundError struct {
	Domain string
}

func (e DomainNamedResourceNotFoundError) Error() string {
	return fmt.Sprintf("No Hosted Zones found for domain %s", e.Domain)
}

type NamedResourceNotFoundError struct {
	Name string
}

func (e NamedResourceNotFoundError) Error() string {
	return fmt.Sprintf("The resource was not found: %s", e.Name)
}
