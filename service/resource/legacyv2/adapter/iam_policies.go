package adapter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/aws-operator/service/keyv2"
	"github.com/giantswarm/microerror"
)

// template related to this adapter: service/templates/cloudformation/iam_policies.yaml

type iamPoliciesAdapter struct {
	WorkerRoleName    string
	WorkerPolicyName  string
	WorkerProfileName string
	KMSKeyARN         string
	S3Bucket          string
}

func (i *iamPoliciesAdapter) getIamPolicies(customObject v1alpha1.AWSConfig, clients Clients) error {
	clusterID := keyv2.ClusterID(customObject)

	i.WorkerPolicyName = keyv2.PolicyName(customObject, prefixWorker)
	i.WorkerProfileName = keyv2.InstanceProfileName(customObject, prefixWorker)
	i.WorkerRoleName = keyv2.RoleName(customObject, prefixWorker)

	// KMSKeyARN
	keyAlias := fmt.Sprintf("alias/%s", clusterID)
	input := &kms.DescribeKeyInput{
		KeyId: aws.String(keyAlias),
	}
	output, err := clients.KMS.DescribeKey(input)
	if err != nil {
		return microerror.Mask(err)
	}
	i.KMSKeyARN = *output.KeyMetadata.Arn

	// S3Bucket
	accountID, err := AccountID(clients)
	if err != nil {
		return microerror.Mask(err)
	}
	i.S3Bucket = keyv2.BucketName(customObject, accountID)

	return nil
}
