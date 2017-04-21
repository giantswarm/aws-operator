package api

import (
	"net"
)

type API struct {
	// AltNames is the alternative names used to generate certificates for the
	// Kubernetes API server.
	AltNames string `json:"alt_names" yaml:"altNames"`
	// ClusterIpRange is the value for command line flag
	// --service-cluster-ip-range of the Kubernetes API server, e.g. 10.0.3.0/24.
	ClusterIPRange string `json:"cluster_ip_range" yaml:"clusterIPRange"`
	// Domain is the API domain of the Kubernetes cluster, e.g.
	// api.<cluster-id>.g8s.fra-1.giantswarm.io.
	Domain       string `json:"domain" yaml:"domain"`
	InsecurePort int    `json:"insecure_port" yaml:"insecurePort"`
	// IP is the Kubernetes API IP, e.g. 172.29.0.1.
	IP         net.IP `json:"ip" yaml:"ip"`
	SecurePort int    `json:"secure_port" yaml:"securePort"`
}
