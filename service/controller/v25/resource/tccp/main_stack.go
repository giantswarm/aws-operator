package tccp

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v25/adapter"
	"github.com/giantswarm/aws-operator/service/controller/v25/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v25/key"
	"github.com/giantswarm/aws-operator/service/controller/v25/templates"
)

func (r *Resource) getMainGuestTemplateBody(ctx context.Context, customObject v1alpha1.AWSConfig, stackState StackState) (string, error) {
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
