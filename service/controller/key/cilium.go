package key

import (
	apiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const (
	CiliumForceDisableKubeProxyAnnotation = "cilium-force-disable-kube-proxy-replacement"
)

func ForceDisableCiliumKubeProxyReplacement(cluster apiv1beta1.Cluster) bool {
	v, found := cluster.Annotations[CiliumForceDisableKubeProxyAnnotation]

	return found && v == "true"
}
