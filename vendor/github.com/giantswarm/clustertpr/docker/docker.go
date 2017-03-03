package docker

type Docker struct {
	Daemon         Daemon   `json:"daemon" yaml:"daemon"`
	Registry       Registry `json:"registry" yaml:"registry"`
	ImageNamespace string   `json:"imageNamespace" yaml:"imageNamespace"`
}
