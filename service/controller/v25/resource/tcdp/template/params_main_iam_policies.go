package template

type ParamsMainIAMPolicies struct {
	EC2ServiceDomain string
	KMSKeyARN        string
	NodePool         ParamsMainIAMPoliciesNodePool
	RegionARN        string
	S3Bucket         string
}

type ParamsMainIAMPoliciesNodePool struct {
	ID string
}
