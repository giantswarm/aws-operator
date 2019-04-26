package template

import (
	"strings"
	"testing"
)

func Test_Controller_Resource_CPI_Template_Render(t *testing.T) {
	var err error

	var params *ParamsMain
	{
		iamRoles := &ParamsMainIAMRoles{
			PeerAccessRoleName: "PeerAccessRoleName",
			Tenant: ParamsMainIAMRolesTenant{
				AWS: ParamsMainIAMRolesTenantAWS{
					Account: ParamsMainIAMRolesTenantAWSAccount{
						ID: "TenantAWSAccountID",
					},
				},
			},
		}

		params = &ParamsMain{
			IAMRoles: iamRoles,
		}
	}

	var templateBody string
	{
		templateBody, err = Render(params)
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
