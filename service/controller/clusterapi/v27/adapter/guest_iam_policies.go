package adapter

import (
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/legacykey"
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
	clusterID := legacykey.ClusterID(cfg.CustomObject)

	i.ClusterID = clusterID
	i.EC2ServiceDomain = legacykey.EC2ServiceDomain(cfg.CustomObject)
	i.MasterPolicyName = legacykey.PolicyNameMaster(cfg.CustomObject)
	i.MasterProfileName = legacykey.ProfileName(cfg.CustomObject, legacykey.KindMaster)
	i.MasterRoleName = legacykey.RoleName(cfg.CustomObject, legacykey.KindMaster)
	i.WorkerPolicyName = legacykey.PolicyNameWorker(cfg.CustomObject)
	i.WorkerProfileName = legacykey.ProfileName(cfg.CustomObject, legacykey.KindWorker)
	i.WorkerRoleName = legacykey.RoleName(cfg.CustomObject, legacykey.KindWorker)
	i.RegionARN = legacykey.RegionARN(cfg.CustomObject)
	i.KMSKeyARN = cfg.TenantClusterKMSKeyARN
	i.S3Bucket = legacykey.BucketName(cfg.CustomObject, cfg.TenantClusterAccountID)

	return nil
}
