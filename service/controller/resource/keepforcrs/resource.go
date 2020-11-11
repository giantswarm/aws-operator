package keepforcrs

import (
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	Name = "keepforcrs"
)

type Config struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
	// NewObjFunc is to return an instance of a pointer for the CR type that
	// should be considered for keeping finalizers.
	//
	//     &infrastructurev1alpha2.AWSControlPlane{}
	//     &infrastructurev1alpha2.AWSMachineDeployment{}
	//
	NewObjFunc func() runtime.Object
}

// Resource receives the runtime object of the underlying controller it is wired
// into and keeps finalizers for that very controller in case the configured
// runtime objects do still exist. This is to have a reliable deletion for the
// following CRs.
//
//     watch         |    delete
//     ---------------------------------------
//     AWSCluster    |    AWSControlPlane
//     AWSCluster    |    AWSMachineDeployment
//
type Resource struct {
	k8sClient  k8sclient.Interface
	logger     micrologger.Logger
	newObjFunc func() runtime.Object
}

func New(config Config) (*Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.NewObjFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.NewObjFunc must not be empty", config)
	}

	r := &Resource{
		k8sClient:  config.K8sClient,
		logger:     config.Logger,
		newObjFunc: config.NewObjFunc,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
