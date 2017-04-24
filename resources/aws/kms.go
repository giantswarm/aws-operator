package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	microerror "github.com/giantswarm/microkit/error"
)

type KMSKey struct {
	Name string
	arn  string
	AWSEntity
}

func (kk KMSKey) fullAlias() string {
	return fmt.Sprintf("alias/%s", kk.Name)
}

func (kk *KMSKey) CreateIfNotExists() (bool, error) {
	return false, fmt.Errorf("KMS keys cannot be reused")
}

func (kk *KMSKey) CreateOrFail() error {
	key, err := kk.Clients.KMS.CreateKey(&kms.CreateKeyInput{})
	if err != nil {
		return microerror.MaskAny(err)
	}

	if _, err := kk.Clients.KMS.CreateAlias(&kms.CreateAliasInput{
		// Alias names need to start from "alias/" prefix.
		AliasName:   aws.String(kk.fullAlias()),
		TargetKeyId: key.KeyMetadata.Arn,
	}); err != nil {
		return microerror.MaskAny(err)
	}

	kk.arn = *key.KeyMetadata.Arn

	return nil
}

func (kk *KMSKey) Delete() error {
	key, err := kk.Clients.KMS.DescribeKey(&kms.DescribeKeyInput{
		KeyId: aws.String(kk.fullAlias()),
	})
	if err != nil {
		return microerror.MaskAny(err)
	}

	if _, err := kk.Clients.KMS.DeleteAlias(&kms.DeleteAliasInput{
		AliasName: aws.String(kk.fullAlias()),
	}); err != nil {
		return microerror.MaskAny(err)
	}

	// AWS API doesn't allow to delete the KMS key immediately, but we can schedule its deletion
	if _, err := kk.Clients.KMS.ScheduleKeyDeletion(&kms.ScheduleKeyDeletionInput{
		KeyId:               key.KeyMetadata.KeyId,
		PendingWindowInDays: aws.Int64(7),
	}); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (kk KMSKey) Arn() string {
	return kk.arn
}
