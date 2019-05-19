package tccp

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/adapter"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/detection"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/key"
)

const (
	// Name is the identifier of the resource.
	Name = "tccpv27"
)

const (
	namedIAMCapability = "CAPABILITY_NAMED_IAM"

	// versionBundleVersionParameterKey is the key name of the Cloud Formation
	// parameter that sets the version bundle version.
	versionBundleVersionParameterKey = "VersionBundleVersionParameter"
)

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
	VPCPeerID                  string
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
	vpcPeerID          string
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
	if config.VPCPeerID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.VPCPeerID must not be empty", config)
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
		vpcPeerID:          config.VPCPeerID,
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
func (r *Resource) searchMasterInstanceID(ctx context.Context, cr v1alpha1.Cluster) (string, error) {
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

func (r *Resource) terminateMasterInstance(ctx context.Context, cr v1alpha1.Cluster) error {
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
