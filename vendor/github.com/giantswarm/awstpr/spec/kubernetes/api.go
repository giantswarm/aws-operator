package kubernetes

type API struct {
	// AltNames is the alternative names used to generate certificates for the
	// Kubernetes API server.
	AltNames string `json:"alt_names"`
	// ClusterIpRange is the value for command line flag
	// --service-cluster-ip-range of the Kubernetes API server, e.g.
	// 172.31.0.0/24.
	ClusterIPRange string `json:"cluster_ip_range"`
	// Domain is the API domain of the Kubernetes cluster, e.g.
	// api.<cluster-id>.g8s.fra-1.giantswarm.io.
	Domain       string `json:"domain"`
	InsecurePort string `json:"insecure_port"`
	// IP is the Kubernetes API IP, e.g. 172.31.0.1,
	IP         string `json:"ip"`
	SecurePort string `json:"secure_port"`
}
