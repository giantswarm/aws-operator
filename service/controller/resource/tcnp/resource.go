package tcnp

import (
	"encoding/json"

	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/v16/service/controller/resource/tcnp/template"
	"github.com/giantswarm/aws-operator/v16/service/internal/changedetection"
	"github.com/giantswarm/aws-operator/v16/service/internal/cloudtags"
	"github.com/giantswarm/aws-operator/v16/service/internal/encrypter"
	"github.com/giantswarm/aws-operator/v16/service/internal/images"
	"github.com/giantswarm/aws-operator/v16/service/internal/recorder"
)

const (
	// Name is the identifier of the resource.
	Name = "tcnp"
)

type Config struct {
	CloudTags cloudtags.Interface
	Detection *changedetection.TCNP
	Encrypter encrypter.Interface
	Event     recorder.Interface
	Images    images.Interface
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger

	AlikeInstances   string
	InstallationName string
}

// Resource implements the TCNP resource, which stands for Tenant Cluster Data
// Plane. We manage a dedicated Cloud Formation stack for each node pool.
type Resource struct {
	cloudtags cloudtags.Interface
	detection *changedetection.TCNP
	encrypter encrypter.Interface
	event     recorder.Interface
	images    images.Interface
	k8sClient k8sclient.Interface
	logger    micrologger.Logger

	alikeInstances   map[string][]template.LaunchTemplateOverride
	installationName string
}

func New(config Config) (*Resource, error) {
	if config.CloudTags == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CloudTags must not be empty", config)
	}
	if config.Detection == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Detection must not be empty", config)
	}
	if config.Encrypter == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Encrypter must not be empty", config)
	}
	if config.Event == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Event must not be empty", config)
	}
	if config.Images == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Images must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.AlikeInstances == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.AlikeInstances must not be empty", config)
	}
	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationName must not be empty", config)
	}

	var alikeInstances map[string][]template.LaunchTemplateOverride
	{
		err := json.Unmarshal([]byte(config.AlikeInstances), &alikeInstances)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	r := &Resource{
		cloudtags: config.CloudTags,
		detection: config.Detection,
		encrypter: config.Encrypter,
		event:     config.Event,
		images:    config.Images,
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		alikeInstances:   alikeInstances,
		installationName: config.InstallationName,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
