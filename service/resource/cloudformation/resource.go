package cloudformation

import (
	"context"

	"github.com/giantswarm/certificatetpr"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"k8s.io/client-go/kubernetes"
)

const (
	// Name is the identifier of the resource.
	Name = "cloudformation"
)

// Config represents the configuration used to create a new config map resource.
type Config struct {
	// Dependencies.
	CertWatcher certificatetpr.Searcher
	//CloudConfig *cloudconfig.CloudConfig
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new config map
// resource by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		CertWatcher: nil,
		//CloudConfig: nil,
		K8sClient: nil,
		Logger:    nil,
	}
}

// Resource implements the config map resource.
type Resource struct {
	// Dependencies.
	certWatcher certificatetpr.Searcher
	//cloudConfig *cloudconfig.CloudConfig
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

// New creates a new configured config map resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.CertWatcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.CertWatcher must not be empty")
	}
	//if config.CloudConfig == nil {
	//	return nil, microerror.Maskf(invalidConfigError, "config.CloudConfig must not be empty")
	//}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	newService := &Resource{
		// Dependencies.
		certWatcher: config.CertWatcher,
		//cloudConfig: config.CloudConfig,
		k8sClient: config.K8sClient,
		logger: config.Logger.With(
			"resource", Name,
		),
	}

	return newService, nil
}

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	// TODO fetch a comparable current state of the cloudformation associated with
	// the processed custom object.
	return nil, nil
}

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	// TODO compute a comparable desired state of the cloudformation associated
	// with the processed custom object.
	return nil, nil
}

func (r *Resource) GetCreateState(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	// TODO compute a comparable create state of the cloudformation associated
	// with the processed custom object. This can be derived by comparing the
	// current and the desired state.
	return nil, nil
}

func (r *Resource) GetDeleteState(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	// TODO compute a comparable delete state of the cloudformation associated
	// with the processed custom object. This can be derived by comparing the
	// current and the desired state.
	return nil, nil
}

func (r *Resource) GetUpdateState(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, interface{}, interface{}, error) {
	// TODO compute a comparable create, delete and update state of the
	// cloudformation associated with the processed custom object. This can be
	// derived by comparing the current and the desired state.
	return nil, nil, nil, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ProcessCreateState(ctx context.Context, obj, createState interface{}) error {
	// TODO process a comparable create state of the cloudformation associated
	// with the processed custom object. This can be done by creating the given
	// create state in some remote API via some configured client.
	return nil
}

func (r *Resource) ProcessDeleteState(ctx context.Context, obj, deleteState interface{}) error {
	// TODO process a comparable delete state of the cloudformation associated
	// with the processed custom object. This can be done by deleting the given
	// delete state in some remote API via some configured client.
	return nil
}

func (r *Resource) ProcessUpdateState(ctx context.Context, obj, updateState interface{}) error {
	// TODO process a comparable update state of the cloudformation associated
	// with the processed custom object. This can be done by updating the given
	// update state in some remote API via some configured client.
	return nil
}

func (r *Resource) Underlying() framework.Resource {
	return r
}
