package elb

// IdleTimeoutSeconds contains the Idle Timeout in seconds of the elastic load balancers.
type IdleTimeoutSeconds struct {
	// API is the Idle Timeout in seconds for the Kubernetes API Server load balancer.
	API int `json:"api" yaml:"api"`
	// Etcd is the Idle Timeout in seconds for the etcd load balancer.
	Etcd int `json:"etcd" yaml:"etcd"`
	// Ingress is the Idle Timeout in seconds for the Ingress load balancer, used for customer traffic.
	Ingress int `json:"ingress" yaml:"ingress"`
}
