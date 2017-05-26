package hostedzones

// HostedZones contains the Hosted Zone IDs of the cluster.
type HostedZones struct {
	// API is the Hosted Zone ID for the Kubernetes API.
	API string `json:"api" yaml:"api"`
	// Etcd is the Hosted Zone ID for the etcd cluster.
	Etcd string `json:"etcd" yaml:"etcd"`
	// Ingress is the Hosted Zone ID for the Ingress resource, used for customer traffic.
	Ingress string `json:"ingress" yaml:"ingress"`
}
