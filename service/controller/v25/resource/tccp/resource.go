package tccp

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/v25/adapter"
	"github.com/giantswarm/aws-operator/service/controller/v25/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v25/detection"
	"github.com/giantswarm/aws-operator/service/controller/v25/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/v25/key"
)

const (
	// Name is the identifier of the resource.
	Name = "tccpv25"
)

const (
	namedIAMCapability = "CAPABILITY_NAMED_IAM"

	// versionBundleVersionParameterKey is the key name of the Cloud Formation
	// parameter that sets the version bundle version.
	versionBundleVersionParameterKey = "VersionBundleVersionParameter"
)

type AWSConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	SessionToken    string
	Region          string
	accountID       string
}

// Config represents the configuration used to create a new cloudformation
// resource.
type Config struct {
	APIWhitelist adapter.APIWhitelist
	// EncrypterRoleManager manages role encryption. This can be supported by
	// different implementations and thus is optional.
	EncrypterRoleManager encrypter.RoleManager
	Logger               micrologger.Logger

	Detection                  *detection.Detection
	EncrypterBackend           string
	GuestPrivateSubnetMaskBits int
	GuestPublicSubnetMaskBits  int
	InstallationName           string
	InstanceMonitoring         bool
	PublicRouteTables          string
	Route53Enabled             bool
}

// Resource implements the cloudformation resource.
type Resource struct {
	apiWhiteList         adapter.APIWhitelist
	encrypterRoleManager encrypter.RoleManager
	logger               micrologger.Logger

	encrypterBackend   string
	detection          *detection.Detection
	installationName   string
	instanceMonitoring bool
	publicRouteTables  string
	route53Enabled     bool
}

// New creates a new configured cloudformation resource.
func New(config Config) (*Resource, error) {
	if config.Detection == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Detection must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.EncrypterBackend == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.EncrypterBackend must not be empty", config)
	}

	r := &Resource{
		apiWhiteList:         config.APIWhitelist,
		detection:            config.Detection,
		encrypterRoleManager: config.EncrypterRoleManager,
		logger:               config.Logger,

		encrypterBackend:   config.EncrypterBackend,
		installationName:   config.InstallationName,
		instanceMonitoring: config.InstanceMonitoring,
		publicRouteTables:  config.PublicRouteTables,
		route53Enabled:     config.Route53Enabled,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) searchMasterInstanceID(ctx context.Context, cr v1alpha1.AWSConfig) (string, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}

	var instanceID string
	{
		i := &ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("tag:Name"),
					Values: []*string{
						aws.String(key.MasterInstanceName(cr)),
					},
				},
				{
					Name: aws.String("tag:giantswarm.io/cluster"),
					Values: []*string{
						aws.String(key.ClusterID(cr)),
					},
				},
			},
		}

		o, err := cc.Client.TenantCluster.AWS.EC2.DescribeInstances(i)
		if err != nil {
			return "", microerror.Mask(err)
		}

		if len(o.Reservations) == 0 {
			return "", microerror.Maskf(notExistsError, "master instance")
		}
		if len(o.Reservations) != 1 {
			return "", microerror.Maskf(executionFailedError, "expected one master instance, got %d", len(o.Reservations))
		}
		if len(o.Reservations[0].Instances) != 1 {
			return "", microerror.Maskf(executionFailedError, "expected one master instance, got %d", len(o.Reservations[0].Instances))
		}

		instanceID = *o.Reservations[0].Instances[0].InstanceId
	}

	return instanceID, nil
}

func (r *Resource) terminateMasterInstance(ctx context.Context, cr v1alpha1.AWSConfig) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var instanceID string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding master instance ID")

		instanceID, err = r.searchMasterInstanceID(ctx, cr)
		if IsNotExists(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find master instance ID")
			r.logger.LogCtx(ctx, "level", "debug", "message", "master instance does not exist")
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found master instance ID %#q", instanceID))
	}

	//{
	//	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("disabling termination protection for master instance %#q", instanceID))
	//
	//	i := &ec2.ModifyInstanceAttributeInput{
	//		DisableApiTermination: &ec2.AttributeBooleanValue{
	//			Value: aws.Bool(false),
	//		},
	//		InstanceId: aws.String(instanceID),
	//	}
	//
	//	_, err = cc.Client.TenantCluster.AWS.EC2.ModifyInstanceAttribute(i)
	//	if err != nil {
	//		return microerror.Mask(err)
	//	}
	//
	//	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("disabled termination protection for master instance %#q", instanceID))
	//}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("terminating master instance %#q", instanceID))

		i := &ec2.TerminateInstancesInput{
			InstanceIds: []*string{
				aws.String(instanceID),
			},
		}

		_, err := cc.Client.TenantCluster.AWS.EC2.TerminateInstances(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("terminated master instance %#q", instanceID))
	}

	return nil
}
