package spec

import "github.com/giantswarm/clustertpr/spec/docker"

type Docker struct {
	Daemon docker.Daemon `json:"daemon" yaml:"daemon"`
}
