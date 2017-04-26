package operator

import (
	"github.com/giantswarm/clustertpr/operator/certctl"
	"github.com/giantswarm/clustertpr/operator/k8svm"
	"github.com/giantswarm/clustertpr/operator/kubectl"
	"github.com/giantswarm/clustertpr/operator/networksetup"
)

type Operator struct {
	Certctl      certctl.Certctl
	K8sVM        k8svm.K8sVM
	Kubectl      kubectl.Kubectl
	NetworkSetup networksetup.NetworkSetup
}
