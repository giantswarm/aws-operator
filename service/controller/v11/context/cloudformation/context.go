package cloudformation

import (
	"context"

	"github.com/giantswarm/microerror"

	cloudformationservice "github.com/giantswarm/aws-operator/service/controller/v11/cloudformation"
)

type contextKey string

const cloudFormationKey contextKey = "cloudformation"

func NewContext(ctx context.Context, cloudformation cloudformationservice.CloudFormation) context.Context {
	return context.WithValue(ctx, cloudFormationKey, &cloudformation)
}

func FromContext(ctx context.Context) (*cloudformationservice.CloudFormation, error) {
	clients, ok := ctx.Value(cloudFormationKey).(*cloudformationservice.CloudFormation)
	if !ok {
		return nil, microerror.Mask(serviceNotFound)
	}

	return clients, nil
}
