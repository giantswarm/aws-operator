package cpi

import (
	"strings"
	"testing"

	"github.com/giantswarm/aws-operator/service/controller/v24/adapter"
	"github.com/giantswarm/aws-operator/service/controller/v24/key"
	"github.com/giantswarm/aws-operator/service/controller/v24/templates"
)

func Test_Resource_CPI_Template(t *testing.T) {
	var err error

	var cpi *adapter.CPI
	{
		c := adapter.CPIConfig{
			PeerAccessRoleName: "PeerAccessRoleName",
			TenantAWSAccountID: "TenantAWSAccountID",
		}

		cpi, err = adapter.NewCPI(c)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	var templateBody string
	{
		templateBody, err = templates.Render(key.CloudFormationHostPreTemplates(), cpi)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	{
		expected := "RoleName: PeerAccessRoleName"
		if !strings.Contains(templateBody, expected) {
			t.Fatal("expected", "match", "got", "none")
		}
	}

	{
		expected := "AWS: 'TenantAWSAccountID'"
		if !strings.Contains(templateBody, expected) {
			t.Fatal("expected", "match", "got", "none")
		}
	}
}
