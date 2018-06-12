package provider

import (
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type AWSConfig struct {
	GuestFramework *framework.Guest
	HostFramework  *framework.Host
	Logger         micrologger.Logger

	ClusterID string
}

type AWS struct {
	guestFramework *framework.Guest
	hostFramework  *framework.Host
	logger         micrologger.Logger

	clusterID string
}

func NewAWS(config AWSConfig) (*AWS, error) {
	if config.GuestFramework == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.GuestFramework must not be empty", config)
	}
	if config.HostFramework == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HostFramework must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.ClusterID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterID must not be empty", config)
	}

	a := &AWS{
		guestFramework: config.GuestFramework,
		hostFramework:  config.HostFramework,
		logger:         config.Logger,

		clusterID: config.ClusterID,
	}

	return a, nil
}

func (a *AWS) AddWorker() error {
	customObject, err := a.hostFramework.AWSCluster(a.clusterID)
	if err != nil {
		return microerror.Mask(err)
	}

	patches := []framework.PatchSpec{
		{
			Op:    "add",
			Path:  "/spec/aws/workers/-",
			Value: customObject.Spec.AWS.Workers[0],
		},
	}

	err = a.hostFramework.ApplyAWSConfigPatch(patches, a.clusterID)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (a *AWS) NumMasters() (int, error) {
	customObject, err := a.hostFramework.AWSCluster(a.clusterID)
	if err != nil {
		return 0, microerror.Mask(err)
	}

	num := len(customObject.Spec.AWS.Masters)

	return num, nil
}

func (a *AWS) NumWorkers() (int, error) {
	customObject, err := a.hostFramework.AWSCluster(a.clusterID)
	if err != nil {
		return 0, microerror.Mask(err)
	}

	num := len(customObject.Spec.AWS.Workers)

	return num, nil
}

func (a *AWS) RemoveWorker() error {
	patches := []framework.PatchSpec{
		{
			Op:   "remove",
			Path: "/spec/aws/workers/1",
		},
	}

	err := a.hostFramework.ApplyAWSConfigPatch(patches, a.clusterID)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (a *AWS) WaitForNodesUp(num int) error {
	err := a.guestFramework.WaitForNodesUp(num)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
