package tccp

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/internal/changedetection"
	"github.com/giantswarm/aws-operator/service/controller/internal/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

const (
	// Name is the identifier of the resource.
	Name = "tccp"
)

const (
	// namedIAMCapability is the AWS specific capability necessary to work with
	// our Cloud Formation templates. It is required for creating worker policy
	// IAM roles.
	namedIAMCapability = "CAPABILITY_NAMED_IAM"
)

// Config represents the configuration used to create a new cloudformation
// resource.
type Config struct {
	// EncrypterRoleManager manages role encryption. This can be supported by
	// different implementations and thus is optional.
	EncrypterRoleManager encrypter.RoleManager
	G8sClient            versioned.Interface
	Logger               micrologger.Logger

	APIWhitelist       APIWhitelist
	CIDRBlockAWSCNI    string
	Detection          *changedetection.TCCP
	InstallationName   string
	InstanceMonitoring bool
	PublicRouteTables  string
	Route53Enabled     bool
}

// Resource implements the cloudformation resource.
type Resource struct {
	encrypterRoleManager encrypter.RoleManager
	g8sClient            versioned.Interface
	logger               micrologger.Logger

	apiWhiteList       APIWhitelist
	cidrBlockAWSCNI    string
	detection          *changedetection.TCCP
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
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.APIWhitelist.Private.Enabled && config.APIWhitelist.Private.SubnetList == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.APIWhitelist.Private.SubnetList must not be empty when %T.APIWhitelist.Private is enabled", config)
	}
	if config.APIWhitelist.Public.Enabled && config.APIWhitelist.Public.SubnetList == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.APIWhitelist.Public.SubnetList must not be empty when %T.APIWhitelist.Public is enabled", config)
	}

	if config.CIDRBlockAWSCNI == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.CIDRBlockAWSCNI must not be empty", config)
	}

	r := &Resource{
		g8sClient:            config.G8sClient,
		detection:            config.Detection,
		encrypterRoleManager: config.EncrypterRoleManager,
		logger:               config.Logger,

		apiWhiteList:       config.APIWhitelist,
		cidrBlockAWSCNI:    config.CIDRBlockAWSCNI,
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

// searchMasterInstanceID tries to find any "active" master instance. The method
// ignores instances that are shutting down or are already terminated. This is
// because we only need to find the master instance in order to terminate it
// before updating the TCCP Cloud Formation stack. In case the master instance
// is already terminated, we ignore it. The used filter name is the following.
//
//     instance-state-name
//
// To be precise, the following instance states are considered as filter values.
//
//     pending, running, stopping, stopped
//
func (r *Resource) searchMasterInstanceID(ctx context.Context, cr infrastructurev1alpha2.AWSCluster) (string, error) {
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
					Name: aws.String(fmt.Sprintf("tag:%s", key.TagCluster)),
					Values: []*string{
						aws.String(key.ClusterID(&cr)),
					},
				},
				{
					Name: aws.String("instance-state-name"),
					Values: []*string{
						aws.String(ec2.InstanceStateNamePending),
						aws.String(ec2.InstanceStateNameRunning),
						aws.String(ec2.InstanceStateNameStopped),
						aws.String(ec2.InstanceStateNameStopping),
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

func (r *Resource) stopMasterInstance(ctx context.Context, cr infrastructurev1alpha2.AWSCluster) error {
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

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("requesting to stop instance %#q", instanceID))

		i := &ec2.StopInstancesInput{
			InstanceIds: []*string{
				aws.String(instanceID),
			},
		}

		_, err := cc.Client.TenantCluster.AWS.EC2.StopInstances(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("requested to stop instance %#q", instanceID))
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waiting for instance %#q to be stopped", instanceID))

		i := &ec2.DescribeInstancesInput{
			InstanceIds: []*string{
				aws.String(instanceID),
			},
		}

		err := cc.Client.TenantCluster.AWS.EC2.WaitUntilInstanceStopped(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waited for instance %#q to be stopped", instanceID))
	}

	return nil
}

func (r *Resource) terminateMasterInstance(ctx context.Context, cr infrastructurev1alpha2.AWSCluster) error {
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

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("disabling termination protection for master instance %#q", instanceID))

		i := &ec2.ModifyInstanceAttributeInput{
			DisableApiTermination: &ec2.AttributeBooleanValue{
				Value: aws.Bool(false),
			},
			InstanceId: aws.String(instanceID),
		}

		_, err = cc.Client.TenantCluster.AWS.EC2.ModifyInstanceAttribute(i)
		if IsAlreadyTerminated(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not disable termination protection for master instance %#q", instanceID))
			r.logger.LogCtx(ctx, "level", "debug", "message", "master instance is already terminated")
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("disabled termination protection for master instance %#q", instanceID))
	}

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
