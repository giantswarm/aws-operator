package template

type ParamsMainIAMPolicies struct {
	ClusterID            string
	EC2ServiceDomain     string
	HostedZoneID         string
	InternalHostedZoneID string
	KMSKeyARN            string
	RegionARN            string
	S3Bucket             string
	Route53Enabled       bool
}
