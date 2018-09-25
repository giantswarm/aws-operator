package controllercontext

import (
	"context"

	"github.com/giantswarm/microerror"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
	awsservice "github.com/giantswarm/aws-operator/service/aws"
	"github.com/giantswarm/aws-operator/service/controller/v16patch1/cloudconfig"
	cloudformationservice "github.com/giantswarm/aws-operator/service/controller/v16patch1/cloudformation"
	"github.com/giantswarm/aws-operator/service/controller/v16patch1/ebs"
)

type contextKey string

const controllerKey contextKey = "controller"

type Context struct {
	AWSClient      awsclient.Clients
	AWSService     awsservice.Interface
	CloudConfig    cloudconfig.Interface
	CloudFormation cloudformationservice.CloudFormation
	EBSService     ebs.Interface

	// Status holds the data used to communicate between controller's
	// resources. It can be edited in place as Context is stored as
	// a pointer within context.Context.
	Status Status
}

func NewContext(ctx context.Context, c Context) context.Context {
	return context.WithValue(ctx, controllerKey, &c)
}

func FromContext(ctx context.Context) (*Context, error) {
	c, ok := ctx.Value(controllerKey).(*Context)
	if !ok {
		return nil, microerror.Maskf(notFoundError, "context key %q of type %T", controllerKey, controllerKey)
	}

	return c, nil
}
