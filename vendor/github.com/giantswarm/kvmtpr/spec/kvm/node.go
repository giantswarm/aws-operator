package kvm

type Node struct {
	CPUs   int     `json:"cpus" yaml:"cpus"`
	Disk   float64 `json:"disk" yaml:"disk"`
	Memory string  `json:"memory" yaml:"memory"`
}
