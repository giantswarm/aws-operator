package cloudformation

import (
	"bytes"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	awsCF "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/aws-operator/service/key"
	"github.com/giantswarm/awstpr"
)

func newMainStack(customObject awstpr.CustomObject) (*awsCF.CreateStackInput, error) {
	stackName := key.MainStackName(customObject)

	mainCF := &awsCF.CreateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(MainTemplate),
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
