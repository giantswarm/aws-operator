package adapter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v13/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/v13/key"
)

// The template related to this adapter can be found in the following import.
//
//     github.com/giantswarm/aws-operator/service/controller/v13/templates/cloudformation/guest/iam_policies.go
//

type iamPoliciesAdapter struct {
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

func (i *iamPoliciesAdapter) getIamPolicies(cfg Config) error {
	clusterID := key.ClusterID(cfg.CustomObject)

	i.EC2ServiceDomain = key.EC2ServiceDomain(cfg.CustomObject)
	i.MasterPolicyName = key.PolicyName(cfg.CustomObject, prefixMaster)
	i.MasterProfileName = key.InstanceProfileName(cfg.CustomObject, prefixMaster)
	i.MasterRoleName = key.RoleName(cfg.CustomObject, prefixMaster)
	i.WorkerPolicyName = key.PolicyName(cfg.CustomObject, prefixWorker)
	i.WorkerProfileName = key.InstanceProfileName(cfg.CustomObject, prefixWorker)
	i.WorkerRoleName = key.RoleName(cfg.CustomObject, prefixWorker)
	i.RegionARN = key.RegionARN(cfg.CustomObject)

	// KMSKeyARN
	if cfg.EncrypterBackend == encrypter.KMSBackend {
		keyAlias := fmt.Sprintf("alias/%s", clusterID)
		input := &kms.DescribeKeyInput{
			KeyId: aws.String(keyAlias),
		}
		output, err := cfg.Clients.KMS.DescribeKey(input)
		if err != nil {
			return microerror.Mask(err)
		}
		i.KMSKeyARN = *output.KeyMetadata.Arn
	}

	// S3Bucket
	accountID, err := AccountID(cfg.Clients)
	if err != nil {
		return microerror.Mask(err)
	}
	i.S3Bucket = key.BucketName(cfg.CustomObject, accountID)

	return nil
}
