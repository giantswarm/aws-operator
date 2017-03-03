package operator

type Operator struct {
	CertctlVersion      string `json:"certctlVersion" yaml:"certctlVersion"`
	K8sVmVersion        string `json:"k8sVmVersion" yaml:"k8sVmVersion"`
	KubectlVersion      string `json:"kubectlVersion" yaml:"kubectlVersion"`
	NetworkSetupVersion string `json:"networkSetupVersion" yaml:"networkSetupVersion"`
}
