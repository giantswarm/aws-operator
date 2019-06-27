package loadtest

import (
	"context"
)

const (
	// ApdexPassThreshold is the minimum value allowed for the test to pass.
	ApdexPassThreshold = 0.95
	AppChartName       = "loadtest-app-chart"
	CNRAddress         = "https://quay.io"
	CNROrganization    = "giantswarm"
	ChartChannel       = "stable"
	ChartNamespace     = "e2e-app"
	JobChartName       = "stormforger-cli-chart"
	LoadTestNamespace  = "loadtest"
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
