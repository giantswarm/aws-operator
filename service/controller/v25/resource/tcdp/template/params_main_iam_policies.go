package template

type ParamsMainIAMPolicies struct {
	Cluster          ParamsMainIAMPoliciesCluster
	EC2ServiceDomain string
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
