package docker

type Daemon struct {
	CIDR      string `json:"cidr" yaml:"cidr"`
	ExtraArgs string `json:"extraArgs" yaml:"extraArgs"`
}
