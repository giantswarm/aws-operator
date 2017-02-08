package docker

type Docker struct {
	Registry Registry `json:"registry"`
	Daemon   Daemon   `json:"daemon"`
}
