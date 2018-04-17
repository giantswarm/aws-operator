package v1

import (
	"github.com/giantswarm/certs/legacy"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/resource/metricsresource"
	"github.com/giantswarm/randomkeytpr"
	"k8s.io/client-go/kubernetes"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/awsconfig/v1/cloudconfig"
	legacyv2 "github.com/giantswarm/aws-operator/service/awsconfig/v1/resource/legacy"
	"github.com/giantswarm/aws-operator/service/awsconfig/v2/key"
)

const (
	ResourceRetries uint64 = 3
)

type ResourceSetConfig struct {
	CertsSearcher      *legacy.Service
	GuestAWSConfig     awsclient.Config
	HostAWSConfig      awsclient.Config
	K8sClient          kubernetes.Interface
	Logger             micrologger.Logger
	RandomkeysSearcher *randomkeytpr.Service

	HandledVersionBundles []string
	InstallationName      string
	ProjectName           string
	PubKeyFile            string
}

func NewResourceSet(config ResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	if config.CertsSearcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.CertsSearcher must not be empty")
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.RandomkeysSearcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.RandomkeysSearcher must not be empty")
	}
	// GuestAWSConfig is validated in controller.go and this resource set is legacy.
	// HostAWSConfig is validated in controller.go and this resource set is legacy.
	if len(config.HandledVersionBundles) == 0 {
		return nil, microerror.Maskf(invalidConfigError, "config.HandledVersionBundles must not be empty")
	}
	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.InstallationName must not be empty")
	}
	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.ProjectName must not be empty")
	}

	var cloudConfig *cloudconfig.CloudConfig
	{
		c := cloudconfig.Config{
			Logger: config.Logger,
		}

		cloudConfig, err = cloudconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var legacyResource controller.CRUDResourceOps
	{
		legacyConfig := legacyv2.DefaultConfig()
		legacyConfig.AwsConfig = config.GuestAWSConfig
		legacyConfig.AwsHostConfig = config.HostAWSConfig
		legacyConfig.CertWatcher = config.CertsSearcher
		legacyConfig.CloudConfig = cloudConfig
		legacyConfig.InstallationName = config.InstallationName
		legacyConfig.K8sClient = config.K8sClient
		legacyConfig.KeyWatcher = config.RandomkeysSearcher
		legacyConfig.Logger = config.Logger
		legacyConfig.PubKeyFile = config.PubKeyFile

		legacyResource, err = legacyv2.New(legacyConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resources []controller.Resource
	ops := []controller.CRUDResourceOps{
		legacyResource,
	}
	for _, o := range ops {
		c := controller.CRUDResourceConfig{
			Logger: config.Logger,
			Ops:    o,
		}

		r, err := controller.NewCRUDResource(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		resources = append(resources, r)
	}

	// Wrap resources with retry and metrics.
	{
		c := metricsresource.WrapConfig{
			Name: config.ProjectName,
		}

		resources, err = metricsresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	handlesFunc := func(obj interface{}) bool {
		awsConfig, err := key.ToCustomObject(obj)
		if err != nil {
			return false
		}
		versionBundleVersion := key.VersionBundleVersion(awsConfig)

		for _, v := range config.HandledVersionBundles {
			if versionBundleVersion == v {
				return true
			}
		}

		return false
	}

	var resourceSet *controller.ResourceSet
	{
		c := controller.ResourceSetConfig{
			Handles:   handlesFunc,
			Logger:    config.Logger,
			Resources: resources,
		}

		resourceSet, err = controller.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resourceSet, nil
}
