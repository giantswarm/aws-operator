package docker

import (
	"github.com/giantswarm/clustertpr/docker/daemon"
	"github.com/giantswarm/clustertpr/docker/registry"
)

type Docker struct {
	Daemon         daemon.Daemon     `json:"daemon" yaml:"daemon"`
	ImageNamespace string            `json:"imageNamespace" yaml:"imageNamespace"`
	Registry       registry.Registry `json:"registry" yaml:"registry"`
}
