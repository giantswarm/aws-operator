package adapter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/keyv2"
)

// template related to this adapter: service/templates/cloudformation/guest/iam_policies.yaml

type iamPoliciesAdapter struct {
	MasterRoleName    string
	MasterPolicyName  string
	MasterProfileName string
	WorkerRoleName    string
	WorkerPolicyName  string
	WorkerProfileName string
	KMSKeyARN         string
	S3Bucket          string
}

func (i *iamPoliciesAdapter) getIamPolicies(cfg Config) error {
	clusterID := keyv2.ClusterID(cfg.CustomObject)

	i.MasterPolicyName = keyv2.PolicyName(cfg.CustomObject, prefixMaster)
	i.MasterProfileName = keyv2.InstanceProfileName(cfg.CustomObject, prefixMaster)
	i.MasterRoleName = keyv2.RoleName(cfg.CustomObject, prefixMaster)
	i.WorkerPolicyName = keyv2.PolicyName(cfg.CustomObject, prefixWorker)
	i.WorkerProfileName = keyv2.InstanceProfileName(cfg.CustomObject, prefixWorker)
	i.WorkerRoleName = keyv2.RoleName(cfg.CustomObject, prefixWorker)

	// KMSKeyARN
	keyAlias := fmt.Sprintf("alias/%s", clusterID)
	input := &kms.DescribeKeyInput{
		KeyId: aws.String(keyAlias),
	}
	output, err := cfg.Clients.KMS.DescribeKey(input)
	if err != nil {
		return microerror.Mask(err)
	}
	i.KMSKeyARN = *output.KeyMetadata.Arn

	// S3Bucket
	accountID, err := AccountID(cfg.Clients)
	if err != nil {
		return microerror.Mask(err)
	}
	i.S3Bucket = keyv2.BucketName(cfg.CustomObject, accountID)

	return nil
}
