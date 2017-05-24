package node

type Node struct {
	CPUs   int    `json:"cpus" yaml:"cpus"`
	ID     string `json:"id" yaml:"id"`
	Memory string `json:"memory" yaml:"memory"`
}
