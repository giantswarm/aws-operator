package loadtest

import (
	"github.com/giantswarm/helmclient"
	"k8s.io/client-go/kubernetes"
)

// Clients are configured in the e2e test.
type Clients struct {
	ControlPlaneHelmClient helmclient.Interface
	ControlPlaneK8sClient  kubernetes.Interface
}

// LoadTestApp passes values to the loadtest-app chart.
type LoadTestApp struct {
	Ingress LoadTestAppIngress `json:"ingress"`
}

type LoadTestAppIngress struct {
	Hosts []string `json:"hosts"`
}

// LoadTestResults parses results from the Storm Forger API.
type LoadTestResults struct {
	Data LoadTestResultsData `json:"data"`
}

type LoadTestResultsData struct {
	Attributes LoadTestResultsDataAttributes `json:"attributes"`
}

type LoadTestResultsDataAttributes struct {
	BasicStatistics LoadTestResultsDataAttributesBasicStatistics `json:"basic_statistics"`
}

type LoadTestResultsDataAttributesBasicStatistics struct {
	Apdex75 float32 `json:apdex_75`
}

// LoadTestValues passes values to the stormforger-cli chart.
type LoadTestValues struct {
	Auth LoadTestValuesAuth `json:"auth"`
	Test LoadTestValuesTest `json:"test"`
}

type LoadTestValuesAuth struct {
	Token string `json:"token"`
}

type LoadTestValuesTest struct {
	Endpoint string `json:"endpoint"`
}
