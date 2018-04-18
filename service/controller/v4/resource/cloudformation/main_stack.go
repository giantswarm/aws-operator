package cloudformation

import (
	"strconv"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v4/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v4/key"
	"github.com/giantswarm/aws-operator/service/controller/v4/resource/cloudformation/adapter"
	"github.com/giantswarm/aws-operator/service/controller/v4/templates"
)

func newMainStack(customObject v1alpha1.AWSConfig) (StackState, error) {
	stackName := key.MainGuestStackName(customObject)
	workerCount := key.WorkerCount(customObject)

	var workerInstanceType string

	imageID, err := key.ImageID(customObject)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}

	// FIXME: the instance type should not depend on the number of workers.
	// issue: https://github.com/giantswarm/awstpr/issues/47
	if workerCount > 0 {
		workerInstanceType = key.WorkerInstanceType(customObject)
	}

	var masterInstanceType string
	if len(customObject.Spec.AWS.Masters) > 0 {
		masterInstanceType = key.MasterInstanceType(customObject)
	}

	masterCloudConfigVersion := cloudconfig.MasterCloudConfigVersion
	workerCloudConfigVersion := cloudconfig.WorkerCloudConfigVersion

	mainCF := StackState{
		Name:                     stackName,
		MasterImageID:            imageID,
		MasterInstanceType:       masterInstanceType,
		MasterCloudConfigVersion: masterCloudConfigVersion,
		WorkerCount:              strconv.Itoa(workerCount),
		WorkerImageID:            imageID,
		WorkerInstanceType:       workerInstanceType,
		WorkerCloudConfigVersion: workerCloudConfigVersion,
	}

	return mainCF, nil
}

func (r *Resource) getMainGuestTemplateBody(customObject v1alpha1.AWSConfig) (string, error) {
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
