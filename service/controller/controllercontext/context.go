package controllercontext

import (
	"context"

	"github.com/giantswarm/microerror"
)

type contextKey string

const controllerKey contextKey = "controller"

type Context struct {
	// Client holds the client implementations used for several tenant cluster
	// specific actions.
	Client ContextClient
	// Spec holds the desired state of resource during reconciliations. It can be edited
	// in place as Context is stored as a pointer within context.Context.
	Spec ContextSpec
	// Status holds the current state of resources during reconciliation. It
	// can be edited in place as Context is stored as a pointer within context.Context.
	Status ContextStatus
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
