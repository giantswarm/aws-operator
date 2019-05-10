package adapter

import (
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/key"
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
	WorkerRoleName    string
	WorkerPolicyName  string
	WorkerProfileName string
}

func (i *GuestIAMPoliciesAdapter) Adapt(cfg Config) error {
	clusterID := key.ClusterID(cfg.CustomObject)

	i.ClusterID = clusterID
	i.EC2ServiceDomain = key.EC2ServiceDomain(cfg.CustomObject)
	i.MasterPolicyName = key.PolicyNameMaster(cfg.CustomObject)
	i.MasterProfileName = key.ProfileNameMaster(cfg.CustomObject)
	i.MasterRoleName = key.RoleNameMaster(cfg.CustomObject)
	i.WorkerPolicyName = key.PolicyNameWorker(cfg.CustomObject)
	i.WorkerProfileName = key.ProfileNameWorker(cfg.CustomObject)
	i.WorkerRoleName = key.RoleNameWorker(cfg.CustomObject)
	i.RegionARN = key.RegionARN(cfg.CustomObject)
	i.KMSKeyARN = cfg.TenantClusterKMSKeyARN
	i.S3Bucket = key.BucketName(cfg.CustomObject, cfg.TenantClusterAccountID)

	return nil
}
