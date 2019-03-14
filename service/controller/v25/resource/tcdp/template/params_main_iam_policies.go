package template

type ParamsMainIAMPolicies struct {
	ClusterID         string
	EC2ServiceDomain  string
	KMSKeyARN         string
	MasterRoleName    string
	MasterPolicyName  string
	MasterProfileName string
	RegionARN         string
	S3Bucket          string
	WorkerRoleName    string
	WorkerPolicyName  string
	WorkerProfileName string
}
