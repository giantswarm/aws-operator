package node

type Node struct {
	Hostname string `json:"hostname" yaml:"hostname"`
	Memory   string `json:"memory" yaml:"memory"`
	CPUs     int    `json:"cpus" yaml:"cpus"`
}
