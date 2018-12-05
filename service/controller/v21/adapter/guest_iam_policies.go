package adapter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v21/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/v21/key"
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
