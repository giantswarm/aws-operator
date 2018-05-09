package aws

import (
	"context"

	"github.com/giantswarm/aws-operator/client/aws"
)

type contextKey string

const awsClientsKey contextKey = "clients"

func NewContext(ctx context.Context, clients aws.Clients) context.Context {
	return context.WithValue(ctx, awsClientsKey, &clients)
}

func FromContext(ctx context.Context) (*aws.Clients, bool) {
	clients, ok := ctx.Value(awsClientsKey).(*aws.Clients)
	return clients, ok
}
