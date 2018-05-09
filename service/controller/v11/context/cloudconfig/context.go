package cloudconfig

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v11/cloudconfig"
)

type contextKey string

const cloudConfigKey contextKey = "cloudconfig"

func NewContext(ctx context.Context, c cloudconfig.Interface) context.Context {
	return context.WithValue(ctx, cloudConfigKey, c)
}

func FromContext(ctx context.Context) (cloudconfig.Interface, error) {
	clients, ok := ctx.Value(cloudConfigKey).(cloudconfig.Interface)
	if !ok {
		return nil, microerror.Mask(cloudConfigNotFound)
	}

	return clients, nil
}
