package adapter

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v12/key"
)

// The template related to this adapter can be found in the following import.
//
//     github.com/giantswarm/aws-operator/service/controller/v12/templates/cloudformation/guest/iam_policies.go
//

type iamPoliciesAdapter struct {
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

func (i *iamPoliciesAdapter) getIamPolicies(cfg Config) error {
	// clusterID := key.ClusterID(cfg.CustomObject)

	i.MasterPolicyName = key.PolicyName(cfg.CustomObject, prefixMaster)
	i.MasterProfileName = key.InstanceProfileName(cfg.CustomObject, prefixMaster)
	i.MasterRoleName = key.RoleName(cfg.CustomObject, prefixMaster)
	i.WorkerPolicyName = key.PolicyName(cfg.CustomObject, prefixWorker)
	i.WorkerProfileName = key.InstanceProfileName(cfg.CustomObject, prefixWorker)
	i.WorkerRoleName = key.RoleName(cfg.CustomObject, prefixWorker)
	i.RegionARN = key.RegionARN(cfg.CustomObject)

	// KMSKeyARN
	/*
		keyAlias := fmt.Sprintf("alias/%s", clusterID)
		input := &kms.DescribeKeyInput{
			KeyId: aws.String(keyAlias),
		}
		output, err := cfg.Clients.KMS.DescribeKey(input)
		if err != nil {
			return microerror.Mask(err)
		}
		i.KMSKeyARN = *output.KeyMetadata.Arn
	*/

	// S3Bucket
	accountID, err := AccountID(cfg.Clients)
	if err != nil {
		return microerror.Mask(err)
	}
	i.S3Bucket = key.BucketName(cfg.CustomObject, accountID)

	return nil
}
