package cloudformation

import (
	"context"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v11/adapter"
	awsclientcontext "github.com/giantswarm/aws-operator/service/controller/v11/context/awsclient"
	"github.com/giantswarm/aws-operator/service/controller/v11/key"
	"github.com/giantswarm/aws-operator/service/controller/v11/templates"
)

func (r *Resource) getMainGuestTemplateBody(ctx context.Context, customObject v1alpha1.AWSConfig, stackState StackState) (string, error) {
	hostAccountID, err := adapter.AccountID(*r.hostClients)
	if err != nil {
		return "", microerror.Mask(err)
	}

	awsClients, err := awsclientcontext.FromContext(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}

	adapterClients := adapter.Clients{
		CloudFormation: awsClients.CloudFormation,
		EC2:            awsClients.EC2,
		IAM:            awsClients.IAM,
		KMS:            awsClients.KMS,
		ELB:            awsClients.ELB,
	}

	cfg := adapter.Config{
		APIWhitelist: adapter.APIWhitelist{
			Enabled:    r.apiWhiteList.Enabled,
			SubnetList: r.apiWhiteList.SubnetList,
		},
		CustomObject:     customObject,
		Clients:          adapterClients,
		HostClients:      *r.hostClients,
		InstallationName: r.installationName,
		HostAccountID:    hostAccountID,
		StackState: adapter.StackState{
			Name: stackState.Name,

			MasterImageID:              stackState.MasterImageID,
			MasterInstanceResourceName: stackState.MasterInstanceResourceName,
			MasterInstanceType:         stackState.MasterInstanceType,
			MasterCloudConfigVersion:   stackState.MasterCloudConfigVersion,

			WorkerCount:              stackState.WorkerCount,
			WorkerImageID:            stackState.WorkerImageID,
			WorkerInstanceType:       stackState.WorkerInstanceType,
			WorkerCloudConfigVersion: stackState.WorkerCloudConfigVersion,

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

func (r *Resource) getMainHostPreTemplateBody(ctx context.Context, customObject v1alpha1.AWSConfig) (string, error) {
	awsClients, err := awsclientcontext.FromContext(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}

	adapterClients := adapter.Clients{
		CloudFormation: awsClients.CloudFormation,
		EC2:            awsClients.EC2,
		IAM:            awsClients.IAM,
		KMS:            awsClients.KMS,
		ELB:            awsClients.ELB,
	}

	guestAccountID, err := adapter.AccountID(adapterClients)
	if err != nil {
		return "", microerror.Mask(err)
	}
	cfg := adapter.Config{
		CustomObject:   customObject,
		GuestAccountID: guestAccountID,
	}
	adp, err := adapter.NewHostPre(cfg)
	if err != nil {
		return "", microerror.Mask(err)
	}

	rendered, err := templates.Render(key.CloudFormationHostPreTemplates(), adp)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return rendered, nil
}

func (r *Resource) getMainHostPostTemplateBody(ctx context.Context, customObject v1alpha1.AWSConfig) (string, error) {
	awsClients, err := awsclientcontext.FromContext(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}

	adapterClients := adapter.Clients{
		CloudFormation: awsClients.CloudFormation,
		EC2:            awsClients.EC2,
		IAM:            awsClients.IAM,
		KMS:            awsClients.KMS,
		ELB:            awsClients.ELB,
	}

	cfg := adapter.Config{
		CustomObject: customObject,
		Clients:      adapterClients,
		HostClients:  *r.hostClients,
	}
	adp, err := adapter.NewHostPost(cfg)
	if err != nil {
		return "", microerror.Mask(err)
	}

	rendered, err := templates.Render(key.CloudFormationHostPostTemplates(), adp)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return rendered, nil
}
