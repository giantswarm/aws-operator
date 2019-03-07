package template

import (
	"strings"
	"testing"
)

func Test_Template_CloudFormation_CPI(t *testing.T) {
	var err error

	var cpi *CPI
	{
		iamRoles := &CPIIAMRoles{
			PeerAccessRoleName: "PeerAccessRoleName",
			Tenant: CPIIAMRolesTenant{
				AWS: CPIIAMRolesTenantAWS{
					Account: CPIIAMRolesTenantAWSAccount{
						ID: "TenantAWSAccountID",
					},
				},
			},
		}

		cpi = &CPI{
			IAMRoles: iamRoles,
		}
	}

	var templateBody string
	{
		templateBody, err = Render(cpi)
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
