package cloudformation

import (
	"bytes"
	"strconv"
	"text/template"

	"github.com/giantswarm/aws-operator/service/key"
	"github.com/giantswarm/awstpr"
)

func newMainStack(customObject awstpr.CustomObject) (StackState, error) {
	stackName := key.MainStackName(customObject)
	workers := len(customObject.Spec.AWS.Workers)
	var imageID string
	// FIXME: the imageID should nnot depend on the number of workers.
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

func getMainTemplateBody(customObject awstpr.CustomObject) (string, error) {
	t, err := template.New("main").Parse(MainTemplate)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, customObject); err != nil {
		return "", err
	}

	return tpl.String(), nil
}
