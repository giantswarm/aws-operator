package context

import (
	"context"

	"github.com/giantswarm/microerror"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
	awsservice "github.com/giantswarm/aws-operator/service/aws"
	"github.com/giantswarm/aws-operator/service/controller/v12/cloudconfig"
	cloudformationservice "github.com/giantswarm/aws-operator/service/controller/v12/cloudformation"
	"github.com/giantswarm/aws-operator/service/controller/v12/ebs"
)

type contextKey string

const serviceKey contextKey = "service"

type Context struct {
	AWSClient      awsclient.Clients
	AWSService     awsservice.Interface
	CloudConfig    cloudconfig.Interface
	CloudFormation cloudformationservice.CloudFormation
	EBSService     ebs.Interface
}

func NewContext(ctx context.Context, c Context) context.Context {
	return context.WithValue(ctx, serviceKey, &c)
}

func FromContext(ctx context.Context) (*Context, error) {
	c, ok := ctx.Value(serviceKey).(*Context)
	if !ok {
		return nil, microerror.Mask(serviceNotFound)
	}

	return c, nil
}
