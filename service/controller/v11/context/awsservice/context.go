package awsservice

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/aws"
)

type contextKey string

const awsServiceKey contextKey = "awsservice"

func NewContext(ctx context.Context, service aws.Interface) context.Context {
	return context.WithValue(ctx, awsServiceKey, service)
}

func FromContext(ctx context.Context) (aws.Interface, error) {
	service, ok := ctx.Value(awsServiceKey).(aws.Interface)
	if !ok {
		return nil, microerror.Mask(serviceNotFound)
	}

	return service, nil
}
