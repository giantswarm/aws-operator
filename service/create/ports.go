package create

import "github.com/giantswarm/awstpr"

const (
	calicoBGPNetworkPort = 179
	httpPort             = 80
	httpsPort            = 443
	sshPort              = 22
)

// extractMastersSecurityGroupPorts returns the ports that must be opened on the masters security group.
func extractMastersSecurityGroupPorts(cluster awstpr.CustomObject) []int {
	return []int{
		cluster.Spec.Cluster.Kubernetes.API.SecurePort,
		cluster.Spec.Cluster.Etcd.Port,
		sshPort,
	}
}

// extractMastersSecurityGroupPorts returns the ports that must be opened on the workers security group.
func extractWorkersSecurityGroupPorts(cluster awstpr.CustomObject) []int {
	return []int{
		cluster.Spec.Cluster.Kubernetes.IngressController.InsecurePort,
		cluster.Spec.Cluster.Kubernetes.IngressController.SecurePort,
		cluster.Spec.Cluster.Kubernetes.Kubelet.Port,
		sshPort,
		calicoBGPNetworkPort,
		httpsPort, // TODO Move the https and http ports to a separate security group for Ingress ELB.
		httpPort,
	}
}
