package tcnp

import (
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/aws-operator/pkg/awstags"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/changedetection"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
)

const (
	// Name is the identifier of the resource.
	Name = "tcnpv29"
)

type Config struct {
	CMAClient clientset.Interface
	Detection *changedetection.TCNP
	Logger    micrologger.Logger

	InstallationName string
}

// Resource implements the TCNP resource, which stands for Tenant Cluster Data
// Plane. We manage a dedicated Cloud Formation stack for each node pool.
type Resource struct {
	cmaClient clientset.Interface
	detection *changedetection.TCNP
	logger    micrologger.Logger

	installationName string
}

func New(config Config) (*Resource, error) {
	if config.CMAClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CMAClient must not be empty", config)
	}
	if config.Detection == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Detection must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationName must not be empty", config)
	}

	r := &Resource{
		cmaClient: config.CMAClient,
		detection: config.Detection,
		logger:    config.Logger,

		installationName: config.InstallationName,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) getCloudFormationTags(cr v1alpha1.MachineDeployment) []*cloudformation.Tag {
	tags := key.AWSTags(&cr, r.installationName)
	tags[key.TagStack] = key.StackTCNP
	tags[key.TagMachineDeployment] = key.MachineDeploymentID(&cr)
	return awstags.NewCloudFormation(tags)
}
