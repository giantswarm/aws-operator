package calico

type Calico struct {
	CIDR   string `json:"cidr"`
	MTU    string `json:"mtu"`
	Subnet string `json:"subnet"`
}
