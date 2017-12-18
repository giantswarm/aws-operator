package adapter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/aws-operator/service/keyv2"
	"github.com/giantswarm/microerror"
)

// template related to this adapter: service/templates/cloudformation/worker_policy.yaml

type workerPolicyAdapter struct {
	WorkerRoleName    string
	WorkerPolicyName  string
	WorkerProfileName string
	KMSKeyARN         string
	S3Bucket          string
}

func (w *workerPolicyAdapter) getWorkerPolicy(customObject v1alpha1.AWSConfig, clients Clients) error {
	clusterID := keyv2.ClusterID(customObject)

	w.WorkerPolicyName = keyv2.PolicyName(customObject, prefixWorker)
	w.WorkerProfileName = keyv2.InstanceProfileName(customObject, prefixWorker)
	w.WorkerRoleName = keyv2.RoleName(customObject, prefixWorker)

	// KMSKeyARN
	keyAlias := fmt.Sprintf("alias/%s", clusterID)
	input := &kms.DescribeKeyInput{
		KeyId: aws.String(keyAlias),
	}
	output, err := clients.KMS.DescribeKey(input)
	if err != nil {
		return microerror.Mask(err)
	}
	w.KMSKeyARN = *output.KeyMetadata.Arn

	// S3Bucket
	accountID, err := AccountID(clients)
	if err != nil {
		return microerror.Mask(err)
	}
	w.S3Bucket = keyv2.BucketName(customObject, accountID)

	return nil
}
