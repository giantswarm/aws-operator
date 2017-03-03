package calico

type Calico struct {
	CIDR   string `json:"cidr" yaml:"cidr"`
	MTU    string `json:"mtu" yaml:"mtu"`
	Subnet string `json:"subnet" yaml:"subnet"`
}
