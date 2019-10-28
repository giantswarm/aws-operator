package template

type ParamsMainIAMRoles struct {
	PeerAccessRoleName string
	Tenant             ParamsMainIAMRolesTenant
}

type ParamsMainIAMRolesTenant struct {
	AWS ParamsMainIAMRolesTenantAWS
}

type ParamsMainIAMRolesTenantAWS struct {
	Account ParamsMainIAMRolesTenantAWSAccount
}

type ParamsMainIAMRolesTenantAWSAccount struct {
	ID string
}
