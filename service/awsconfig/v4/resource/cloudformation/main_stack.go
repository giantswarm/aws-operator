package cloudformation

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"text/template"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/awsconfig/v2/key"
	"github.com/giantswarm/aws-operator/service/awsconfig/v2/resource/cloudformation/adapter"
	// NOTE(PK): This import is disturbing. I'm not bothering. It's first candidate to go away.
	"github.com/giantswarm/aws-operator/service/awsconfig/v3/cloudconfig"
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
	hostAccountID, err := adapter.AccountID(*r.HostClients)
	if err != nil {
		return "", microerror.Mask(err)
	}
	cfg := adapter.Config{
		CustomObject:     customObject,
		Clients:          *r.Clients,
		HostClients:      *r.HostClients,
		InstallationName: r.installationName,
		HostAccountID:    hostAccountID,
	}
	adp, err := adapter.NewGuest(cfg)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return r.getMainTemplateBody(cloudFormationGuestTemplatesDirectory, adp)
}

func (r *Resource) getMainHostPreTemplateBody(customObject v1alpha1.AWSConfig) (string, error) {
	guestAccountID, err := adapter.AccountID(*r.Clients)
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

	return r.getMainTemplateBody(cloudFormationHostPreTemplatesDirectory, adp)
}

func (r *Resource) getMainHostPostTemplateBody(customObject v1alpha1.AWSConfig) (string, error) {
	cfg := adapter.Config{
		CustomObject: customObject,
		Clients:      *r.Clients,
		HostClients:  *r.HostClients,
	}
	adp, err := adapter.NewHostPost(cfg)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return r.getMainTemplateBody(cloudFormationHostPostTemplatesDirectory, adp)
}

func (r *Resource) getMainTemplateBody(tplDir string, adp adapter.Adapter) (string, error) {
	main := template.New("")

	var t *template.Template
	var err error

	// parse templates
	baseDir, err := os.Getwd()
	if err != nil {
		return "", microerror.Mask(err)
	}

	rootDir, err := key.RootDir(baseDir, adapter.RootDirElement)
	if err != nil {
		return "", microerror.Mask(err)
	}
	templatesDir := filepath.Join(rootDir, tplDir)

	files, err := ioutil.ReadDir(templatesDir)
	if err != nil {
		return "", microerror.Mask(err)
	}
	templates := []string{}
	for _, file := range files {
		templates = append(templates, filepath.Join(templatesDir, file.Name()))
	}
	t, err = main.ParseFiles(templates...)
	if err != nil {
		return "", microerror.Mask(err)
	}

	var tpl bytes.Buffer
	if err := t.ExecuteTemplate(&tpl, "main", adp); err != nil {
		return "", microerror.Mask(err)
	}

	return tpl.String(), nil
}
