package cloudformation

import (
	"bytes"
	"strconv"
	"text/template"

	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/key"
)

var (
	components = []string{
		MainTemplate,
		LaunchConfigurationTemplate,
		AutoScalingGroupTemplate,
	}
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
	main := template.New("main")

	var t *template.Template
	var err error

	for _, component := range components {
		t, err = main.Parse(component)
		if err != nil {
			return "", microerror.Mask(err)
		}
	}

	var tpl bytes.Buffer

	adaptor := adaptor{}

	if err := adaptor.getMain(customObject, r.awsClients); err != nil {
		return "", microerror.Mask(err)
	}

	if err := t.Execute(&tpl, adaptor); err != nil {
		return "", microerror.Mask(err)
	}

	return tpl.String(), nil
}
