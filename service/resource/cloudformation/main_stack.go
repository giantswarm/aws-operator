package cloudformation

import (
	"bytes"
	"text/template"

	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/aws-operator/service/key"
	"github.com/giantswarm/awstpr"
)

func newMainStack(customObject awstpr.CustomObject) (StackState, error) {
	stackName := key.MainStackName(customObject)

	mainCF := StackState{
		Name:    stackName,
		Outputs: []*awscloudformation.Output{},
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
