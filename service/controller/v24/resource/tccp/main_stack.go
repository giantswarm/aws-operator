package tccp

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v24/adapter"
	"github.com/giantswarm/aws-operator/service/controller/v24/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v24/key"
	"github.com/giantswarm/aws-operator/service/controller/v24/templates"
)

func (r *Resource) getMainGuestTemplateBody(ctx context.Context, customObject v1alpha1.AWSConfig, stackState StackState) (string, error) {
	hostAccountID, err := adapter.AccountID(*r.hostClients)
	if err != nil {
		return "", microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}

	adapterClients := adapter.Clients{
		CloudFormation: cc.AWSClient.CloudFormation,
		EC2:            cc.AWSClient.EC2,
		IAM:            cc.AWSClient.IAM,
		KMS:            cc.AWSClient.KMS,
		ELB:            cc.AWSClient.ELB,
		STS:            cc.AWSClient.STS,
	}

	cfg := adapter.Config{
		APIWhitelist: adapter.APIWhitelist{
			Enabled:    r.apiWhiteList.Enabled,
			SubnetList: r.apiWhiteList.SubnetList,
		},
		CustomObject:      customObject,
		Clients:           adapterClients,
		EncrypterBackend:  r.encrypterBackend,
		HostClients:       *r.hostClients,
		InstallationName:  r.installationName,
		HostAccountID:     hostAccountID,
		PublicRouteTables: r.publicRouteTables,
		Route53Enabled:    r.route53Enabled,
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
