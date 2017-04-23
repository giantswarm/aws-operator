package create

import "github.com/giantswarm/awstpr"

const sshPort = 22

func extractMasterPortsFromTPR(cluster awstpr.CustomObject) []int {
	var ports = []int{
		cluster.Spec.Cluster.Kubernetes.API.SecurePort,
		sshPort,
	}

	return ports
}

func extractWorkerPortsFromTPR(cluster awstpr.CustomObject) []int {
	var ports = []int{
		cluster.Spec.Cluster.Kubernetes.IngressController.InsecurePort,
		cluster.Spec.Cluster.Kubernetes.IngressController.SecurePort,
		sshPort,
	}

	return ports
}
