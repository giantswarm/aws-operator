package cleanuploadbalancers

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	// Name is the identifier of the resource.
	Name = "cleanuploadbalancers"
)

const (
	cloudProviderClusterTagValue = "owned"
	cloudProviderServiceTagKey   = "kubernetes.io/service-name"
	loadBalancerTagChunkSize     = 20
)

// Config represents the configuration used to create a new loadbalancer resource.
type Config struct {
	// Dependencies.
	Logger micrologger.Logger
}

// Resource implements the loadbalancer resource.
type Resource struct {
	// Dependencies.
	logger micrologger.Logger
}

// New creates a new configured loadbalancer resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		// Dependencies.
		logger: config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func splitLoadBalancers(loadBalancerNames []*string, chunkSize int) [][]*string {
	chunks := make([][]*string, 0)

	for i := 0; i < len(loadBalancerNames); i += chunkSize {
		endPos := i + chunkSize

		if endPos > len(loadBalancerNames) {
			endPos = len(loadBalancerNames)
		}

		chunks = append(chunks, loadBalancerNames[i:endPos])
	}

	return chunks
}
