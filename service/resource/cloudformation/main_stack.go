package cloudformation

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"text/template"

	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/key"
)

func newMainStack(customObject awstpr.CustomObject) (StackState, error) {
	stackName := key.MainStackName(customObject)
	workers := len(customObject.Spec.AWS.Workers)
	var imageID string
	// FIXME: the imageID should not depend on the number of workers.
	// issue: https://github.com/giantswarm/awstpr/issues/47
	if workers > 0 {
		imageID = customObject.Spec.AWS.Workers[0].ImageID
	}
	clusterVersion := key.ClusterVersion(customObject)

	mainCF := StackState{
		Name:           stackName,
		Workers:        strconv.Itoa(workers),
		ImageID:        imageID,
		ClusterVersion: clusterVersion,
	}

	return mainCF, nil
}

func (r *Resource) getMainTemplateBody(customObject awstpr.CustomObject) (string, error) {
	main := template.New("")

	var t *template.Template
	var err error

	// parse templates
	baseDir, err := filepath.Abs(filepath.Join("../../../", cloudFormationTemplatesDirectory))
	if err != nil {
		return "", microerror.Mask(err)
	}
	files, err := ioutil.ReadDir(baseDir)
	if err != nil {
		return "", microerror.Mask(err)
	}
	templates := []string{}
	for _, file := range files {
		templates = append(templates, filepath.Join(baseDir, file.Name()))
	}
	t, err = main.ParseFiles(templates...)
	if err != nil {
		return "", microerror.Mask(err)
	}

	adapter, err := newAdapter(customObject, r.awsClients)
	if err != nil {
		return "", microerror.Mask(err)
	}

	var tpl bytes.Buffer
	if err := t.ExecuteTemplate(&tpl, "main", adapter); err != nil {
		return "", microerror.Mask(err)
	}

	return tpl.String(), nil
}
