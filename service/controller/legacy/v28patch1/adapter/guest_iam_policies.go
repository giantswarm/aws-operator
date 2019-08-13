package adapter

import (
	"github.com/giantswarm/aws-operator/service/controller/legacy/v28patch1/key"
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
	i.MasterPolicyName = key.PolicyName(cfg.CustomObject, key.KindMaster)
	i.MasterProfileName = key.InstanceProfileName(cfg.CustomObject, key.KindMaster)
	i.MasterRoleName = key.RoleName(cfg.CustomObject, key.KindMaster)
	i.WorkerPolicyName = key.PolicyName(cfg.CustomObject, key.KindWorker)
	i.WorkerProfileName = key.InstanceProfileName(cfg.CustomObject, key.KindWorker)
	i.WorkerRoleName = key.RoleName(cfg.CustomObject, key.KindWorker)
	i.RegionARN = key.RegionARN(cfg.CustomObject)
	i.KMSKeyARN = cfg.TenantClusterKMSKeyARN
	i.S3Bucket = key.BucketName(cfg.CustomObject, cfg.TenantClusterAccountID)

	return nil
}
