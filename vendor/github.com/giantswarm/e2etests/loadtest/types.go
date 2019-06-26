package loadtest

type LoadTestApp struct {
	Ingress LoadTestAppIngress `json:"ingress"`
}

type LoadTestAppIngress struct {
	Hosts []string `json:"hosts"`
}
