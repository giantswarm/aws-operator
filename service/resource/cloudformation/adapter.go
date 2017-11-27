package cloudformation

import (
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/microerror"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
)

type hydrater func(awstpr.CustomObject, awsutil.Clients) error

type adapter struct {
	ASGType string

	lauchConfigAdapter
	autoScalingGroupAdapter
}

type lauchConfigAdapter struct {
	AssociatePublicIPAddress bool
	BlockDeviceMappings      []BlockDeviceMapping
	IAMInstanceProfileName   string
	ImageID                  string
	InstanceType             string
	SecurityGroupID          string
	SmallCloudConfig         string
}

type BlockDeviceMapping struct {
	DeviceName string
	VolumeSize string
	VolumeType string
}

type autoScalingGroupAdapter struct {
	ASGMinSize             int
	ASGMaxSize             int
	AZ                     string
	HealthCheckGracePeriod string
	LoadBalancerName       string
	MaxBatchSize           string
	MinInstancesInService  int
	RollingUpdatePauseTime string
	SubnetID               string
}

func newAdapter(customObject awstpr.CustomObject, clients awsutil.Clients) (adapter, error) {
	a := adapter{}

	a.ASGType = "worker"

	hydraters := []hydrater{
		a.getAutoScalingGroup,
		a.getLaunchConfiguration,
	}

	for _, h := range hydraters {
		if err := h(customObject, clients); err != nil {
			return adapter{}, microerror.Mask(err)
		}
	}

	return a, nil
}

func (a *adapter) getLaunchConfiguration(customObject awstpr.CustomObject, clients awsutil.Clients) error {
	if len(customObject.Spec.AWS.Workers) == 0 {
		return microerror.Mask(invalidConfigError)
	}

	a.ImageID = customObject.Spec.AWS.Workers[0].ImageID
	a.InstanceType = customObject.Spec.AWS.Workers[0].InstanceType

	return nil
}

func (a *adapter) getAutoScalingGroup(customObject awstpr.CustomObject, clients awsutil.Clients) error {
	a.AZ = customObject.Spec.AWS.AZ

	return nil
}
