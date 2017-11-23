package cloudformation

import (
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/microerror"
	"golang.org/x/sync/errgroup"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
)

type hydrater func(awstpr.CustomObject, awsutil.Clients) error

type adaptor struct {
	ASGType string

	// worker launch configuration
	ImageID                  string
	SecurityGroupID          string
	InstanceType             string
	IAMInstanceProfileName   string
	BlockDeviceMappings      []BlockDeviceMapping
	AssociatePublicIPAddress bool
	SmallCloudConfig         string

	// worker autoscaling group
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

type BlockDeviceMapping struct {
	DeviceName string
	VolumeSize string
	VolumeType string
}

func (a *adaptor) getMain(customObject awstpr.CustomObject, clients awsutil.Clients) error {
	a.ASGType = "worker"

	var g errgroup.Group

	hydraters := []hydrater{
		a.getAutoScalingGroup,
		a.getLaunchConfiguration,
	}

	// each hydrater works on different data and may need to query AWS API, so that
	// they are run in separate goroutines
	for _, h := range hydraters {
		// https://golang.org/doc/faq#closures_and_goroutines
		h := h
		g.Go(func() error {
			return h(customObject, clients)
		})
	}

	if err := g.Wait(); err != nil {
		return microerror.Mask(err)
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
