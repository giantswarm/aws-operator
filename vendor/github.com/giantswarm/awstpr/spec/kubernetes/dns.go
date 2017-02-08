package kubernetes

type DNS struct {
	// IP is the kube DNScluster IP, e.g. 172.31.0.10.
	IP string `json:"ip"`
}
