package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/microerror"
)

type KMSKey struct {
	Name string
	arn  string
	AWSEntity
}

func (kk *KMSKey) CreateIfNotExists() (bool, error) {
	if kk.Name == "" {
		return false, microerror.Mask(kmsKeyAliasEmptyError)
	}

	existingKey, err := kk.findExisting()
	if err != nil {
		return false, microerror.Mask(err)
	}

	if existingKey != nil {
		kk.arn = *existingKey.Arn

		return false, nil
	}

	if err := kk.CreateOrFail(); err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}
func (kk *KMSKey) CreateOrFail() error {
	if kk.Name == "" {
		return microerror.Mask(kmsKeyAliasEmptyError)
	}

	key, err := kk.Clients.KMS.CreateKey(&kms.CreateKeyInput{})
	if err != nil {
		return microerror.Mask(err)
	}

	if _, err := kk.Clients.KMS.CreateAlias(&kms.CreateAliasInput{
		// Alias names need to start from "alias/" prefix.
		AliasName:   aws.String(kk.fullAlias()),
		TargetKeyId: key.KeyMetadata.Arn,
	}); err != nil {
		return microerror.Mask(err)
	}

	kk.arn = *key.KeyMetadata.Arn

	return nil
}

func (kk *KMSKey) Delete() error {
	key, err := kk.Clients.KMS.DescribeKey(&kms.DescribeKeyInput{
		KeyId: aws.String(kk.fullAlias()),
	})
	if err != nil {
		return microerror.Mask(err)
	}

	if _, err := kk.Clients.KMS.DeleteAlias(&kms.DeleteAliasInput{
		AliasName: aws.String(kk.fullAlias()),
	}); err != nil {
		return microerror.Mask(err)
	}

	// AWS API doesn't allow to delete the KMS key immediately, but we can schedule its deletion
	if _, err := kk.Clients.KMS.ScheduleKeyDeletion(&kms.ScheduleKeyDeletionInput{
		KeyId:               key.KeyMetadata.KeyId,
		PendingWindowInDays: aws.Int64(7),
	}); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (kk KMSKey) Arn() string {
	return kk.arn
}

func (kk KMSKey) findExisting() (*kms.KeyMetadata, error) {
	resp, err := kk.Clients.KMS.DescribeKey(&kms.DescribeKeyInput{
		KeyId: aws.String(kk.fullAlias()),
	})
	if err != nil {
		if awserr, ok := err.(awserr.Error); ok && isNotFoundError(awserr.Code()) {
			return nil, nil
		}

		return nil, microerror.Mask(err)
	}

	return resp.KeyMetadata, nil
}

func (kk KMSKey) fullAlias() string {
	return fmt.Sprintf("alias/%s", kk.Name)
}

func isNotFoundError(code string) bool {
	return code == kms.ErrCodeNotFoundException
}
