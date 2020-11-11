package tcnpf

import (
	"github.com/aws/aws-sdk-go/service/cloudformation"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/pkg/awstags"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/internal/recorder"
)

const (
	// Name is the identifier of the resource.
	Name = "tcnpf"
)

type Config struct {
	Event  recorder.Interface
	Logger micrologger.Logger

	InstallationName string
}

// Resource implements the TCNPF resource, which stands for Tenant Cluster Node
// Pool Finalizer. We manage a dedicated CF stack for the VPC Peering
// Connections made between the AWS Control Plane Accounts and the AWS Tenant
// Cluster Accounts.
type Resource struct {
	event  recorder.Interface
	logger micrologger.Logger

	installationName string
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		event:  config.Event,
		logger: config.Logger,

		installationName: config.InstallationName,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) getCloudFormationTags(cr infrastructurev1alpha2.AWSMachineDeployment) []*cloudformation.Tag {
	tags := key.AWSTags(&cr, r.installationName)
	tags[key.TagStack] = key.StackTCNPF
	return awstags.NewCloudFormation(tags)
}
