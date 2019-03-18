package tccp

import (
	"context"

	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/pkg/awstags"
	"github.com/giantswarm/aws-operator/service/controller/v25/adapter"
	"github.com/giantswarm/aws-operator/service/controller/v25/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v25/detection"
	"github.com/giantswarm/aws-operator/service/controller/v25/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/v25/key"
	"github.com/giantswarm/aws-operator/service/controller/v25/templates"
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
	G8sClient            versioned.Interface
	Logger               micrologger.Logger

	AdvancedMonitoringEC2      bool
	Detection                  *detection.Detection
	EncrypterBackend           string
	GuestPrivateSubnetMaskBits int
	GuestPublicSubnetMaskBits  int
	InstallationName           string
	PublicRouteTables          string
	Route53Enabled             bool
}

// Resource implements the cloudformation resource.
type Resource struct {
	apiWhiteList         adapter.APIWhitelist
	encrypterRoleManager encrypter.RoleManager
	g8sClient            versioned.Interface
	logger               micrologger.Logger

	encrypterBackend           string
	detection                  *detection.Detection
	guestPrivateSubnetMaskBits int
	guestPublicSubnetMaskBits  int
	installationName           string
	monitoring                 bool
	publicRouteTables          string
	route53Enabled             bool
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

	// GuestPrivateSubnetMaskBits && GuestPublicSubnetMaskBits has been
	// validated on upper level because all IPAM related configuration
	// information is present there.
	if config.EncrypterBackend == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.EncrypterBackend must not be empty", config)
	}

	r := &Resource{
		apiWhiteList:         config.APIWhitelist,
		detection:            config.Detection,
		encrypterRoleManager: config.EncrypterRoleManager,
		g8sClient:            config.G8sClient,
		logger:               config.Logger,

		encrypterBackend:           config.EncrypterBackend,
		guestPrivateSubnetMaskBits: config.GuestPrivateSubnetMaskBits,
		guestPublicSubnetMaskBits:  config.GuestPublicSubnetMaskBits,
		installationName:           config.InstallationName,
		monitoring:                 config.AdvancedMonitoringEC2,
		publicRouteTables:          config.PublicRouteTables,
		route53Enabled:             config.Route53Enabled,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) getCloudFormationTags(customObject v1alpha1.AWSConfig) []*awscloudformation.Tag {
	tags := key.ClusterTags(customObject, r.installationName)
	return awstags.NewCloudFormation(tags)
}

func (r *Resource) newTemplateBody(ctx context.Context, customObject v1alpha1.AWSConfig, stackState StackState) (string, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}

	cfg := adapter.Config{
		APIWhitelist: adapter.APIWhitelist{
			Enabled:    r.apiWhiteList.Enabled,
			SubnetList: r.apiWhiteList.SubnetList,
		},
		ControlPlaneAccountID:           cc.Status.ControlPlane.AWSAccountID,
		ControlPlaneNATGatewayAddresses: cc.Status.ControlPlane.NATGateway.Addresses,
		ControlPlanePeerRoleARN:         cc.Status.ControlPlane.PeerRole.ARN,
		ControlPlaneVPCCidr:             cc.Status.ControlPlane.VPC.CIDR,
		CustomObject:                    customObject,
		EncrypterBackend:                r.encrypterBackend,
		InstallationName:                r.installationName,
		PublicRouteTables:               r.publicRouteTables,
		Route53Enabled:                  r.route53Enabled,
		StackState: adapter.StackState{
			Name: stackState.Name,

			DockerVolumeResourceName:   stackState.DockerVolumeResourceName,
			MasterImageID:              stackState.MasterImageID,
			MasterInstanceResourceName: stackState.MasterInstanceResourceName,
			MasterInstanceType:         stackState.MasterInstanceType,
			MasterCloudConfigVersion:   stackState.MasterCloudConfigVersion,
			MasterInstanceMonitoring:   stackState.MasterInstanceMonitoring,

			WorkerCloudConfigVersion: stackState.WorkerCloudConfigVersion,
			WorkerDesired:            cc.Status.TenantCluster.TCCP.ASG.DesiredCapacity,
			WorkerDockerVolumeSizeGB: stackState.WorkerDockerVolumeSizeGB,
			WorkerImageID:            stackState.WorkerImageID,
			WorkerInstanceMonitoring: stackState.WorkerInstanceMonitoring,
			WorkerInstanceType:       stackState.WorkerInstanceType,
			WorkerMax:                cc.Status.TenantCluster.TCCP.ASG.MaxSize,
			WorkerMin:                cc.Status.TenantCluster.TCCP.ASG.MinSize,

			VersionBundleVersion: stackState.VersionBundleVersion,
		},
		TenantClusterAccountID: cc.Status.TenantCluster.AWSAccountID,
		TenantClusterKMSKeyARN: cc.Status.TenantCluster.KMS.KeyARN,
	}

	adp, err := adapter.NewGuest(cfg)
	if err != nil {
		return "", microerror.Mask(err)
	}

	rendered, err := templates.Render(key.CloudFormationGuestTemplates(), adp)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return rendered, nil
}

func toCreateStackInput(v interface{}) (awscloudformation.CreateStackInput, error) {
	if v == nil {
		return awscloudformation.CreateStackInput{}, nil
	}

	createStackInput, ok := v.(awscloudformation.CreateStackInput)
	if !ok {
		return awscloudformation.CreateStackInput{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", createStackInput, v)
	}

	return createStackInput, nil
}

func toStackState(v interface{}) (StackState, error) {
	if v == nil {
		return StackState{}, nil
	}

	stackState, ok := v.(StackState)
	if !ok {
		return StackState{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", stackState, v)
	}

	return stackState, nil
}
