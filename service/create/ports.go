package create

import "github.com/giantswarm/awstpr"

const sshPort = 22

func extractPortsFromTPR(cluster awstpr.CustomObject) []int {
	var ports = []int{
		cluster.Spec.Cluster.Kubernetes.API.SecurePort,
		sshPort,
	}

	return ports
}
