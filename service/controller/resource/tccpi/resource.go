package tccpi

import (
	"github.com/aws/aws-sdk-go/service/cloudformation"
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/v13/pkg/awstags"
	"github.com/giantswarm/aws-operator/v13/service/controller/key"
	"github.com/giantswarm/aws-operator/v13/service/internal/recorder"
)

const (
	// Name is the identifier of the resource.
	Name = "tccpi"
)

type Config struct {
	Event  recorder.Interface
	Logger micrologger.Logger

	InstallationName string
}

// Resource implements the CPI resource, which stands for Control Plane
// Initializer. This was formerly known as the host pre stack. We manage a
// dedicated CF stack for the IAM role and VPC Peering setup.
type Resource struct {
	event  recorder.Interface
	logger micrologger.Logger

	installationName string
}

func New(config Config) (*Resource, error) {
	if config.Event == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Event must not be empty", config)
	}
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

func (r *Resource) getCloudFormationTags(cr infrastructurev1alpha3.AWSCluster) []*cloudformation.Tag {
	tags := key.AWSTags(&cr, r.installationName)
	tags[key.TagStack] = key.StackTCCPI
	return awstags.NewCloudFormation(tags)
}
