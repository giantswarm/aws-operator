package cloudformation

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"fmt"
	"github.com/giantswarm/aws-operator/service/controller/v10/adapter"
	"github.com/giantswarm/aws-operator/service/controller/v10/key"
	"github.com/giantswarm/aws-operator/service/controller/v10/templates"
)

func (r *Resource) getMainGuestTemplateBody(customObject v1alpha1.AWSConfig, stackState StackState) (string, error) {
	hostAccountID, err := adapter.AccountID(*r.hostClients)
	if err != nil {
		return "", microerror.Mask(err)
	}
	cfg := adapter.Config{
		CustomObject:     customObject,
		Clients:          *r.clients,
		HostClients:      *r.hostClients,
		InstallationName: r.installationName,
		HostAccountID:    hostAccountID,
		APIWhitelist: adapter.APIWhitelist{
			Enabled:    r.apiWhiteList.Enabled,
			SubnetList: r.apiWhiteList.SubnetList,
		},
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

	fmt.Printf("main_stack.go: whitelist enabled: %t\n", cfg.APIWhitelist.Enabled)
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

func (r *Resource) getMainHostPreTemplateBody(customObject v1alpha1.AWSConfig) (string, error) {
	guestAccountID, err := adapter.AccountID(*r.clients)
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

func (r *Resource) getMainHostPostTemplateBody(customObject v1alpha1.AWSConfig) (string, error) {
	cfg := adapter.Config{
		CustomObject: customObject,
		Clients:      *r.clients,
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
