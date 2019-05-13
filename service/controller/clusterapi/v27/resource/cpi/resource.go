package cpi

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
	Name = "cpiv27"
)

type Config struct {
	Logger micrologger.Logger

	InstallationName string
}

// Resource implements the CPI resource, which stands for Control Plane
// Initializer. This was formerly known as the host pre stack. We manage a
// dedicated CF stack for the IAM role and VPC Peering setup.
type Resource struct {
	logger micrologger.Logger

	installationName string
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		logger: config.Logger,

		installationName: config.InstallationName,
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
