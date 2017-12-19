package legacyv2

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"text/template"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/keyv2"
	"github.com/giantswarm/aws-operator/service/resource/legacyv2/adapter"
)

func newMainStack(customObject v1alpha1.AWSConfig) (StackState, error) {
	stackName := keyv2.MainStackName(customObject)
	workers := len(customObject.Spec.AWS.Workers)
	var imageID string
	// FIXME: the imageID should not depend on the number of workers.
	// issue: https://github.com/giantswarm/awstpr/issues/47
	if workers > 0 {
		imageID = customObject.Spec.AWS.Workers[0].ImageID
	}
	clusterVersion := keyv2.ClusterVersion(customObject)

	mainCF := StackState{
		Name:           stackName,
		Workers:        strconv.Itoa(workers),
		ImageID:        imageID,
		ClusterVersion: clusterVersion,
	}

	return mainCF, nil
}

func (r *Resource) getMainTemplateBody(customObject v1alpha1.AWSConfig) (string, error) {
	main := template.New("")

	var t *template.Template
	var err error

	// parse templates
	baseDir, err := os.Getwd()
	if err != nil {
		return "", microerror.Mask(err)
	}

	rootDir, err := keyv2.RootDir(baseDir, adapter.RootDirElement)
	if err != nil {
		return "", microerror.Mask(err)
	}
	templatesDir := filepath.Join(rootDir, cloudFormationTemplatesDirectory)

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

	adapter, err := adapter.New(customObject, *r.awsClients)
	if err != nil {
		return "", microerror.Mask(err)
	}

	var tpl bytes.Buffer
	if err := t.ExecuteTemplate(&tpl, "main", adapter); err != nil {
		return "", microerror.Mask(err)
	}

	return tpl.String(), nil
}
