package asg

import "context"

// Interface describes how implementations should behave when providing
// information about ASGs. Note that when caching enabled the returned results
// stay consistently the same throughout a reconciliation loop.
type Interface interface {
	// Drainable returns any drainable ASG name found. When using this method the
	// returned ASG name will be the first one found having an active lifecycle
	// hook configured. E.g. when using an implementation configured for HA
	// Masters. Note that there may be one or three masters.
	Drainable(ctx context.Context, obj interface{}) (string, error)
}
