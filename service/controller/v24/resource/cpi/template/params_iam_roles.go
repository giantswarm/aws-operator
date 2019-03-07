package template

type MainParamsIAMRoles struct {
	PeerAccessRoleName string
	Tenant             MainParamsIAMRolesTenant
}

type MainParamsIAMRolesTenant struct {
	AWS MainParamsIAMRolesTenantAWS
}

type MainParamsIAMRolesTenantAWS struct {
	Account MainParamsIAMRolesTenantAWSAccount
}

type MainParamsIAMRolesTenantAWSAccount struct {
	ID string
}
