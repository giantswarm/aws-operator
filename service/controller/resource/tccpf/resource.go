package tccpf

import (
	"github.com/aws/aws-sdk-go/service/cloudformation"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/pkg/awstags"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

const (
	// Name is the identifier of the resource.
	Name = "tccpf"
)

type Config struct {
	Logger micrologger.Logger

	InstallationName string
	Route53Enabled   bool
}

// Resource implements the TCCPF resource, which stands for Tenant Cluster
// Control Plane Finalizer. This was formerly known as the host main stack. We
// manage a dedicated CF stack for the record sets and routing tables setup.
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

	r := &Resource{
		logger: config.Logger,

		installationName: config.InstallationName,
		route53Enabled:   config.Route53Enabled,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) getCloudFormationTags(cr infrastructurev1alpha2.AWSCluster) []*cloudformation.Tag {
	tags := key.AWSTags(&cr, r.installationName)
	tags[key.TagStack] = key.StackTCCPF
	return awstags.NewCloudFormation(tags)
}
