package operator

type Operator struct {
	CertctlVersion      string `json:"certctlVersion"`
	KubectlVersion      string `json:"kubectlVersion"`
	NetworkSetupVersion string `json:"networkSetupVersion"`
}
