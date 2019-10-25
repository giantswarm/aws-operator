package tcnpf

import (
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/pkg/awstags"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

const (
	// Name is the identifier of the resource.
	Name = "tcnpfv31"
)

type Config struct {
	Logger micrologger.Logger

	InstallationName string
}

// Resource implements the TCNPF resource, which stands for Tenant Cluster Node
// Pool Finalizer. We manage a dedicated CF stack for the VPC Peering
// Connections made between the AWS Control Plane Accounts and the AWS Tenant
// Cluster Accounts.
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

func (r *Resource) getCloudFormationTags(cr v1alpha1.MachineDeployment) []*cloudformation.Tag {
	tags := key.AWSTags(&cr, r.installationName)
	tags[key.TagStack] = key.StackTCNPF
	return awstags.NewCloudFormation(tags)
}
