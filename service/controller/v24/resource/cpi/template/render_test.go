package template

import (
	"strings"
	"testing"
)

func Test_Controller_Resource_CPI_Template_Render(t *testing.T) {
	var err error

	var params *MainParams
	{
		iamRoles := &MainParamsIAMRoles{
			PeerAccessRoleName: "PeerAccessRoleName",
			Tenant: MainParamsIAMRolesTenant{
				AWS: MainParamsIAMRolesTenantAWS{
					Account: MainParamsIAMRolesTenantAWSAccount{
						ID: "TenantAWSAccountID",
					},
				},
			},
		}

		params = &MainParams{
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
