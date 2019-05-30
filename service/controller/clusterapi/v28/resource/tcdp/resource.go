package tcdp

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
	Name = "tcdpv28"
)

type Config struct {
	Logger micrologger.Logger

	InstallationName string
}

// Resource implements the TCDP resource, which stands for Tenant Cluster Data
// Plane. We manage a dedicated Cloud Formation stack for each node pool.
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

func (r *Resource) getCloudFormationTags(cluster v1alpha1.Cluster) []*cloudformation.Tag {
	tags := key.ClusterTags(cluster, r.installationName)
	return awstags.NewCloudFormation(tags)
}
