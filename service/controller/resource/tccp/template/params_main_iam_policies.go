package template

type ParamsMainIAMPolicies struct {
	ClusterID         string
	EC2ServiceDomain  string
	KMSKeyARN         string
	MasterRoleName    string
	MasterPolicyName  string
	MasterProfileName string
	RegionARN         string
	Route53Enabled    bool
	S3Bucket          string
}
