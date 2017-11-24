package cloudformation

import (
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/microerror"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
)

type hydrater func(awstpr.CustomObject, awsutil.Clients) error

type adaptor struct {
	ASGType string

	lauchConfigAdaptor
	autoScalingGroupAdaptor
}

type lauchConfigAdaptor struct {
	ImageID                  string
	SecurityGroupID          string
	InstanceType             string
	IAMInstanceProfileName   string
	BlockDeviceMappings      []BlockDeviceMapping
	AssociatePublicIPAddress bool
	SmallCloudConfig         string
}

type BlockDeviceMapping struct {
	DeviceName string
	VolumeSize string
	VolumeType string
}

type autoScalingGroupAdaptor struct {
	SubnetID               string
	AZ                     string
	ASGMinSize             int
	ASGMaxSize             int
	LoadBalancerName       string
	HealthCheckGracePeriod string
	MinInstancesInService  int
	MaxBatchSize           string
	RollingUpdatePauseTime string
}

func (a *adaptor) getMain(customObject awstpr.CustomObject, clients awsutil.Clients) error {
	a.ASGType = "worker"

	hydraters := []hydrater{
		a.getAutoScalingGroup,
		a.getLaunchConfiguration,
	}

	for _, h := range hydraters {
		if err := h(customObject, clients); err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (a *adaptor) getLaunchConfiguration(customObject awstpr.CustomObject, clients awsutil.Clients) error {
	if len(customObject.Spec.AWS.Workers) == 0 {
		return microerror.Mask(invalidConfigError)
	}

	a.ImageID = customObject.Spec.AWS.Workers[0].ImageID
	a.InstanceType = customObject.Spec.AWS.Workers[0].InstanceType

	return nil
}

func (a *adaptor) getAutoScalingGroup(customObject awstpr.CustomObject, clients awsutil.Clients) error {
	a.AZ = customObject.Spec.AWS.Region

	return nil
}
