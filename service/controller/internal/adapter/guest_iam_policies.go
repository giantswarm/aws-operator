package adapter

import (
	"github.com/giantswarm/aws-operator/service/controller/key"
)

type GuestIAMPoliciesAdapter struct {
	ClusterID         string
	EC2ServiceDomain  string
	KMSKeyARN         string
	MasterRoleName    string
	MasterPolicyName  string
	MasterProfileName string
	RegionARN         string
	S3Bucket          string
}

func (i *GuestIAMPoliciesAdapter) Adapt(cfg Config) error {
	clusterID := key.ClusterID(&cfg.CustomObject)

	i.ClusterID = clusterID
	i.EC2ServiceDomain = key.EC2ServiceDomain(cfg.AWSRegion)
	i.MasterPolicyName = key.PolicyNameMaster(cfg.CustomObject)
	i.MasterProfileName = key.ProfileNameMaster(cfg.CustomObject)
	i.MasterRoleName = key.RoleNameMaster(cfg.CustomObject)
	i.RegionARN = key.RegionARN(cfg.AWSRegion)
	i.KMSKeyARN = cfg.TenantClusterKMSKeyARN
	i.S3Bucket = key.BucketName(&cfg.CustomObject, cfg.TenantClusterAccountID)

	return nil
}
