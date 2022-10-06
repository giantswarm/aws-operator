package key

import (
	"github.com/giantswarm/k8smetadata/pkg/annotation"
	apiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

func ForceDisableCiliumKubeProxyReplacement(cluster apiv1beta1.Cluster) bool {
	v, found := cluster.Annotations[annotation.CiliumForceDisableKubeProxyAnnotation]

	return found && v == "true"
}
