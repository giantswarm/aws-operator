package loadtest

import (
	"context"
)

type Interface interface {
	// Test executes the loadtest test using the configured provider
	// implementation. The provider implementation has to be aware of the tenant
	// cluster it has to act against. The test processes the following steps to
	// ensure Nginx Ingress Controller and cluster-autoscaler behave correctly
	// under load.
	//
	//     - Enable HPA for Nginx Ingress Controller using user configmap.
	//     - Install Storm Forger testapp as test workload.
	//     - Wait for test workload to be ready.
	//     - Trigger load test using Storm Forger CLI.
	//     - Analyze results from Storm Forger CLI.
	//
	Test(ctx context.Context) error
}
