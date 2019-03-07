package adapter

type CPIIAMRoles struct {
	PeerAccessRoleName string
	Tenant             CPIIAMRolesTenant
}

type CPIIAMRolesTenant struct {
	AWS CPIIAMRolesTenantAWS
}

type CPIIAMRolesTenantAWS struct {
	Account CPIIAMRolesTenantAWSAccount
}

type CPIIAMRolesTenantAWSAccount struct {
	ID string
}
