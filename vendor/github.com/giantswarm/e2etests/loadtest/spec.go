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
)

type Interface interface {
	// Test executes the loadtest test that checks that tenant cluster
	// components behave correctly under load. This primarily involves testing
	// the HPA configuration for Nginx Ingress Controller is correct and
	// interacts correctly with the cluster-autoscaler when it is enabled.
	//
	// The load test is performed by Stormforger. Their testapp is installed as
	// the test workload and a job is created to trigger the loadtest via their
	// CLI.
	//
	// https://github.com/stormforger/cli
	// https://github.com/stormforger/testapp
	//
	//     - Generate loadtest-app endpoint for the tenant cluster.
	//     - Enable HPA for Nginx Ingress Controller via user configmap.
	//     - Install loadtest-app chart in the tenant cluster.
	//     - Wait for loadtest-app deployment to be ready.
	//     - Install stormforger-cli chart in the tenant cluster.
	//     - Wait for stormforger-cli job to be completed.
	//     - Get logs for stormforger-cli pod with the results.
	//     - Parse the results and determine whether the test passed.
	//
	Test(ctx context.Context) error
}
