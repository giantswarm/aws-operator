package template

type ParamsMainIAMPolicies struct {
	Cluster          ParamsMainIAMPoliciesCluster
	EC2ServiceDomain string
	EnableAWSCNI     bool
	KMSKeyARN        string
	NodePool         ParamsMainIAMPoliciesNodePool
	RegionARN        string
	S3Bucket         string
}

type ParamsMainIAMPoliciesCluster struct {
	ID string
}

type ParamsMainIAMPoliciesNodePool struct {
	ID string
}
