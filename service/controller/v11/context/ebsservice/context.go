package ebsservice

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v11/ebs"
)

type contextKey string

const ebsServiceKey contextKey = "ebsservice"

func NewContext(ctx context.Context, ebsService ebs.Interface) context.Context {
	return context.WithValue(ctx, ebsServiceKey, ebsService)
}

func FromContext(ctx context.Context) (ebs.Interface, error) {
	clients, ok := ctx.Value(ebsServiceKey).(ebs.Interface)
	if !ok {
		return nil, microerror.Mask(serviceNotFound)
	}

	return clients, nil
}
