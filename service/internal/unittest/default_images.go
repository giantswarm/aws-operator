package unittest

import k8scloudconfig "github.com/giantswarm/k8scloudconfig/v10/pkg/template"

func DefaultImages() k8scloudconfig.Images {
	return k8scloudconfig.Images{
		CalicoCNI:                    "1.0.0",
		CalicoKubeControllers:        "1.0.0",
		CalicoNode:                   "1.0.0",
		Etcd:                         "1.0.0",
		Hyperkube:                    "1.0.0",
		KubernetesAPIHealthz:         "9ccdc9dc55a01b1fde2aea73901d0a699909c9cd",
		KubernetesNetworkSetupDocker: "9ccdc9dc55a01b1fde2aea73901d0a699909c9cd",
	}
}
