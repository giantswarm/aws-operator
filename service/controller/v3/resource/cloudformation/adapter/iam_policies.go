package adapter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v3/key"
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
	clusterID := key.ClusterID(cfg.CustomObject)

	i.MasterPolicyName = key.PolicyName(cfg.CustomObject, prefixMaster)
	i.MasterProfileName = key.InstanceProfileName(cfg.CustomObject, prefixMaster)
	i.MasterRoleName = key.RoleName(cfg.CustomObject, prefixMaster)
	i.WorkerPolicyName = key.PolicyName(cfg.CustomObject, prefixWorker)
	i.WorkerProfileName = key.InstanceProfileName(cfg.CustomObject, prefixWorker)
	i.WorkerRoleName = key.RoleName(cfg.CustomObject, prefixWorker)

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
	i.S3Bucket = key.BucketName(cfg.CustomObject, accountID)

	return nil
}
