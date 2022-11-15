package template

type ParamsMainIAMPolicies struct {
	AccountID             string
	AWSBaseDomain         string
	CloudfrontAliasDomain string
	CloudfrontDomain      string
	ClusterID             string
	EC2ServiceDomain      string
	HostedZoneID          string
	InternalHostedZoneID  string
	KMSKeyARN             string
	Region                string
	RegionARN             string
	S3Bucket              string
	Route53Enabled        bool
}
