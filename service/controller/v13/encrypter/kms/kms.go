package kms

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/aws-operator/service/controller/v13/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v13/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/v13/key"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type Encrypter struct {
	logger micrologger.Logger

	installationName string
}

type EncrypterConfig struct {
	Logger micrologger.Logger

	InstallationName string
}

func NewEncrypter(c *EncrypterConfig) (*Encrypter, error) {
	if c.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", c)
	}

	if c.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationName must not be empty", c)
	}

	kms := &Encrypter{
		logger: c.Logger,

		installationName: c.InstallationName,
	}

	return kms, nil
}

func (k *Encrypter) CreateKey(ctx context.Context, customObject v1alpha1.AWSConfig, keyAlias string) error {
	sc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	key, err := sc.AWSClient.KMS.CreateKey(&kms.CreateKeyInput{})
	if err != nil {
		return microerror.Mask(err)
	}

	if _, err := sc.AWSClient.KMS.CreateAlias(&kms.CreateAliasInput{
		AliasName:   aws.String(keyAlias),
		TargetKeyId: key.KeyMetadata.Arn,
	}); err != nil {
		return microerror.Mask(err)
	}

	if _, err := sc.AWSClient.KMS.EnableKeyRotation(&kms.EnableKeyRotationInput{
		KeyId: key.KeyMetadata.KeyId,
	}); err != nil {
		return microerror.Mask(err)
	}

	if _, err := sc.AWSClient.KMS.TagResource(&kms.TagResourceInput{
		KeyId: key.KeyMetadata.KeyId,
		Tags:  k.getKMSTags(customObject),
	}); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (k *Encrypter) DeleteKey(ctx context.Context, keyAlias string) error {
	sc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	// Get the KMS key ID using the key alias.
	key, err := sc.AWSClient.KMS.DescribeKey(&kms.DescribeKeyInput{
		KeyId: aws.String(keyAlias),
	})
	if err != nil {
		return microerror.Mask(err)
	}

	// Delete the key alias.
	if _, err := sc.AWSClient.KMS.DeleteAlias(&kms.DeleteAliasInput{
		AliasName: aws.String(keyAlias),
	}); err != nil {
		return microerror.Mask(err)
	}

	// AWS API doesn't allow to delete the KMS key immediately, but we can schedule its deletion.
	if _, err := sc.AWSClient.KMS.ScheduleKeyDeletion(&kms.ScheduleKeyDeletionInput{
		KeyId:               key.KeyMetadata.KeyId,
		PendingWindowInDays: aws.Int64(pendingDeletionWindow),
	}); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (k *Encrypter) CurrentState(ctx context.Context, customObject v1alpha1.AWSConfig) (encrypter.EncryptionKeyState, error) {
	var currentState encrypter.EncryptionKeyState

	sc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return currentState, microerror.Mask(err)
	}

	clusterID := key.ClusterID(customObject)
	alias := toAlias(clusterID)
	input := &kms.DescribeKeyInput{
		KeyId: aws.String(alias),
	}

	output, err := sc.AWSClient.KMS.DescribeKey(input)
	if IsKeyNotFound(err) {
		// Fall through.
		return currentState, nil
	}
	if err != nil {
		return currentState, microerror.Mask(err)
	}

	currentState.KeyID = *output.KeyMetadata.KeyId
	currentState.KeyName = alias

	return currentState, nil
}

func (k *Encrypter) DesiredState(ctx context.Context, customObject v1alpha1.AWSConfig) (encrypter.EncryptionKeyState, error) {
	desiredState := encrypter.EncryptionKeyState{}

	clusterID := key.ClusterID(customObject)
	desiredState.KeyName = toAlias(clusterID)

	return desiredState, nil
}

func (k *Encrypter) getKMSTags(customObject v1alpha1.AWSConfig) []*kms.Tag {
	clusterTags := key.ClusterTags(customObject, k.installationName)
	kmsTags := []*kms.Tag{}

	for k, v := range clusterTags {
		tag := &kms.Tag{
			TagKey:   aws.String(k),
			TagValue: aws.String(v),
		}

		kmsTags = append(kmsTags, tag)
	}

	return kmsTags
}

func toAlias(keyID string) string {
	return fmt.Sprintf("alias/%s", keyID)
}
