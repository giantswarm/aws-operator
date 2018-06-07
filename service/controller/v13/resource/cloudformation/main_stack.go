package cloudformation

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v13/adapter"
	"github.com/giantswarm/aws-operator/service/controller/v13/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v13/key"
	"github.com/giantswarm/aws-operator/service/controller/v13/templates"
)

func (r *Resource) getMainGuestTemplateBody(ctx context.Context, customObject v1alpha1.AWSConfig, stackState StackState) (string, error) {
	hostAccountID, err := adapter.AccountID(*r.hostClients)
	if err != nil {
		return "", microerror.Mask(err)
	}

	sc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}

	adapterClients := adapter.Clients{
		CloudFormation: sc.AWSClient.CloudFormation,
		EC2:            sc.AWSClient.EC2,
		IAM:            sc.AWSClient.IAM,
		KMS:            sc.AWSClient.KMS,
		ELB:            sc.AWSClient.ELB,
		STS:            sc.AWSClient.STS,
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
		Route53Enabled:   r.route53Enabled,
		StackState: adapter.StackState{
			Name: stackState.Name,

			MasterImageID:              stackState.MasterImageID,
			MasterInstanceResourceName: stackState.MasterInstanceResourceName,
			MasterInstanceType:         stackState.MasterInstanceType,
			MasterCloudConfigVersion:   stackState.MasterCloudConfigVersion,
			MasterInstanceMonitoring:   stackState.MasterInstanceMonitoring,

			WorkerCount:              stackState.WorkerCount,
			WorkerImageID:            stackState.WorkerImageID,
			WorkerInstanceMonitoring: stackState.WorkerInstanceMonitoring,
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
	sc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}

	adapterClients := adapter.Clients{
		CloudFormation: sc.AWSClient.CloudFormation,
		EC2:            sc.AWSClient.EC2,
		IAM:            sc.AWSClient.IAM,
		KMS:            sc.AWSClient.KMS,
		ELB:            sc.AWSClient.ELB,
		STS:            sc.AWSClient.STS,
	}

	guestAccountID, err := adapter.AccountID(adapterClients)
	if err != nil {
		return "", microerror.Mask(err)
	}
	cfg := adapter.Config{
		CustomObject:   customObject,
		GuestAccountID: guestAccountID,
		Route53Enabled: r.route53Enabled,
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

func (r *Resource) getMainHostPostTemplateBody(ctx context.Context, customObject v1alpha1.AWSConfig, guestMainStackState StackState) (string, error) {
	sc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}

	adapterClients := adapter.Clients{
		CloudFormation: sc.AWSClient.CloudFormation,
		EC2:            sc.AWSClient.EC2,
		IAM:            sc.AWSClient.IAM,
		KMS:            sc.AWSClient.KMS,
		ELB:            sc.AWSClient.ELB,
		STS:            sc.AWSClient.STS,
	}

	cfg := adapter.Config{
		CustomObject:   customObject,
		Clients:        adapterClients,
		HostClients:    *r.hostClients,
		Route53Enabled: r.route53Enabled,
		StackState: adapter.StackState{
			HostedZoneNameServers: guestMainStackState.HostedZoneNameServers,
		},
	}
	adp, err := adapter.NewHostPost(cfg)
	if err != nil {
		return "", microerror.Mask(err)
	}

	fmt.Println("================================================================================")
	fmt.Println("================================================================================")
	fmt.Println("================================================================================")
	fmt.Println("================================================================================")
	fmt.Println("================================================================================")
	fmt.Println("adp.Route53Enabled", adp.Route53Enabled)
	fmt.Println("adp.GuestHostedZoneNameServers", adp.GuestHostedZoneNameServers)
	fmt.Println("cfg.StackState.HostedZoneNameServers", cfg.StackState.HostedZoneNameServers)
	fmt.Println("================================================================================")
	fmt.Println("================================================================================")
	fmt.Println("================================================================================")
	fmt.Println("================================================================================")
	fmt.Println("================================================================================")
	fmt.Println("================================================================================")

	rendered, err := templates.Render(key.CloudFormationHostPostTemplates(), adp)
	if err != nil {
		return "", microerror.Mask(err)
	}

	fmt.Println("================================================================================")
	fmt.Println("================================================================================")
	fmt.Println("================================================================================")
	fmt.Println("================================================================================")
	fmt.Println("================================================================================")
	fmt.Println("host post stack:")
	fmt.Println(rendered)
	fmt.Println("================================================================================")
	fmt.Println("================================================================================")
	fmt.Println("================================================================================")
	fmt.Println("================================================================================")
	fmt.Println("================================================================================")
	fmt.Println("================================================================================")

	return rendered, nil
}
