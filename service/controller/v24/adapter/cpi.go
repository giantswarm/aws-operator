package adapter

// CPIConfig represents the config for the adapter collection for the Control
// Plane Initializer management.
type CPIConfig struct {
	PeerAccessRoleName string
	TenantAWSAccountID string
}

// CPI is the adapter collection for the Control Plane Initializer management.
type CPI struct {
	IAMRoles *CPIIAMRoles
}

func NewCPI(config CPIConfig) (*CPI, error) {
	var iamRoles *CPIIAMRoles
	{
		iamRoles = &CPIIAMRoles{
			PeerAccessRoleName: config.PeerAccessRoleName,
			Tenant: CPIIAMRolesTenant{
				AWS: CPIIAMRolesTenantAWS{
					Account: CPIIAMRolesTenantAWSAccount{
						ID: config.TenantAWSAccountID,
					},
				},
			},
		}
	}

	cpi := &CPI{
		IAMRoles: iamRoles,
	}

	return cpi, nil
}
