package tcdp

import (
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/pkg/awstags"
	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
)

const (
	// Name is the identifier of the resource.
	Name = "tcdpv29"
)

type Config struct {
	Logger micrologger.Logger

	InstallationName   string
	InstanceMonitoring bool
}

// Resource implements the TCDP resource, which stands for Tenant Cluster Data
// Plane. We manage a dedicated Cloud Formation stack for each node pool.
type Resource struct {
	logger micrologger.Logger

	installationName   string
	instanceMonitoring bool
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationName must not be empty", config)
	}

	r := &Resource{
		logger: config.Logger,

		installationName:   config.InstallationName,
		instanceMonitoring: config.InstanceMonitoring,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) getCloudFormationTags(cr v1alpha1.MachineDeployment) []*cloudformation.Tag {
	tags := key.ClusterTags(cr, r.installationName)
	tags[label.MachineDeployment] = key.ToMachineDeployment(&cr)
	return awstags.NewCloudFormation(tags)
}
