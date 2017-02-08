package operator

type Operator struct {
	CertctlVersion      string `json:"certctl_ersion"`
	KubectlVersion      string `json:"kubectl_version"`
	NetworkSetupVersion string `json:"networkSetupVersion"`
}
