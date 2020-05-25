package asg

import "context"

type Interface interface {
	// Any returns the first ASG name found. When using this method the returned
	// ASG name will be the only one available. E.g. when using an implementation
	// configured for Node Pools.
	Any(ctx context.Context, obj interface{}) (string, error)
	// Drainable returns any drainable ASG name found. When using this method the
	// returned ASG name will be the first one found having an active lifecycle
	// hook configured. E.g. when using an implementation configured for HA
	// Masters. Note that there may be one or three masters.
	Drainable(ctx context.Context, obj interface{}) (string, error)
}
