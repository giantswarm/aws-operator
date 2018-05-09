package awsclient

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/client/aws"
)

type contextKey string

const awsClientsKey contextKey = "awsclients"

func NewContext(ctx context.Context, clients aws.Clients) context.Context {
	return context.WithValue(ctx, awsClientsKey, &clients)
}

func FromContext(ctx context.Context) (*aws.Clients, error) {
	clients, ok := ctx.Value(awsClientsKey).(*aws.Clients)
	if !ok {
		return nil, microerror.Mask(clientsNotFound)
	}

	return clients, nil
}
