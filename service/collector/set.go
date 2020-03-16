package collector

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/exporterkit/collector"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"

	clientaws "github.com/giantswarm/aws-operator/client/aws"
)

type SetConfig struct {
	G8sClient versioned.Interface
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	AWSConfig             clientaws.Config
	InstallationName      string
	TrustedAdvisorEnabled bool
}

// Set is basically only a wrapper for the operator's collector implementations.
// It eases the initialization and prevents some weird import mess so we do not
// have to alias packages. There is also the benefit of the helper type kept
// private so we do not need to expose this magic.
type Set struct {
	*collector.Set
}

func NewSet(config SetConfig) (*Set, error) {
	var err error

	var h *helper
	{
		c := helperConfig{
			G8sClient: config.G8sClient,
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			AWSConfig: config.AWSConfig,
		}

		h, err = newHelper(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var cfCollector *CloudFormation
	{
		c := CloudFormationConfig{
			Helper: h,
			Logger: config.Logger,

			InstallationName: config.InstallationName,
		}

		cfCollector, err = NewCloudFormation(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var asgCollector *ASG
	{
		c := ASGConfig{
			Helper: h,
			Logger: config.Logger,

			InstallationName: config.InstallationName,
		}

		asgCollector, err = NewASG(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var ec2InstancesCollector *EC2Instances
	{
		c := EC2InstancesConfig{
			Helper: h,
			Logger: config.Logger,

			InstallationName: config.InstallationName,
		}

		ec2InstancesCollector, err = NewEC2Instances(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var elbCollector *ELB
	{
		c := ELBConfig{
			Helper: h,
			Logger: config.Logger,

			InstallationName: config.InstallationName,
		}

		elbCollector, err = NewELB(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var trustedAdvisorCollector *TrustedAdvisor
	{
		c := TrustedAdvisorConfig{
			Helper: h,
			Logger: config.Logger,
		}

		trustedAdvisorCollector, err = NewTrustedAdvisor(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var vpcCollector *VPC
	{
		c := VPCConfig{
			Helper: h,
			Logger: config.Logger,

			InstallationName: config.InstallationName,
		}

		vpcCollector, err = NewVPC(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var collectorSet *collector.Set
	{
		c := collector.SetConfig{
			Collectors: []collector.Interface{
				cfCollector,
				asgCollector,
				ec2InstancesCollector,
				elbCollector,
				vpcCollector,
			},
			Logger: config.Logger,
		}

		if config.TrustedAdvisorEnabled {
			config.Logger.Log("level", "debug", "message", "trusted advisor collector is enabled")
			c.Collectors = append(c.Collectors, trustedAdvisorCollector)
		}

		collectorSet, err = collector.NewSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	s := &Set{
		Set: collectorSet,
	}

	return s, nil
}
