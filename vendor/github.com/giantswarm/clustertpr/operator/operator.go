package operator

import (
	"github.com/giantswarm/clustertpr/operator/certctl"
	"github.com/giantswarm/clustertpr/operator/k8svm"
	"github.com/giantswarm/clustertpr/operator/kubectl"
	"github.com/giantswarm/clustertpr/operator/networksetup"
)

type Operator struct {
	Certctl      certctl.Certctl           `json:"certctl" yaml:"certctl"`
	K8sVM        k8svm.K8sVM               `json:"k8sVM" yaml:"k8sVM"`
	Kubectl      kubectl.Kubectl           `json:"kubectl" yaml:"kubectl"`
	NetworkSetup networksetup.NetworkSetup `json:"networkSetup" yaml:"networkSetup"`
}
