package cpf

import (
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/pkg/awstags"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/key"
)

const (
	// Name is the identifier of the resource.
	Name = "cpfv27"
)

type Config struct {
	Logger micrologger.Logger

	EncrypterBackend string
	InstallationName string
	Route53Enabled   bool
}

// Resource implements the CPF resource, which stands for Control Plane
// Finalizer. This was formerly known as the host post stack. We manage a
// dedicated CF stack for the record sets and routing tables setup.
type Resource struct {
	logger micrologger.Logger

	encrypterBackend string
	installationName string
	route53Enabled   bool
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.EncrypterBackend == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.EncrypterBackend must not be empty", config)
	}

	r := &Resource{
		logger: config.Logger,

		encrypterBackend: config.EncrypterBackend,
		installationName: config.InstallationName,
		route53Enabled:   config.Route53Enabled,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) getCloudFormationTags(cr v1alpha1.Cluster) []*cloudformation.Tag {
	tags := key.ClusterTags(cr, r.installationName)
	return awstags.NewCloudFormation(tags)
}
