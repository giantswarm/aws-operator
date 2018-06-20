package provider

const (
	CNRAddress      = "https://quay.io"
	CNROrganization = "giantswarm"
	ChartChannel    = "stable"
	ChartName       = "e2e-app-chart"
	ChartNamespace  = "e2e-app"
)

type Interface interface {
	InstallTestApp() error
	RebootMaster() error
	ReplaceMaster() error
	WaitForAPIDown() error
	WaitForGuestReady() error
}

type Patch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}
