package kms

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/aws-operator/pkg/awstags"
	"github.com/giantswarm/aws-operator/service/controller/v17/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v17/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/v17/key"
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

func (e *Encrypter) EnsureCreatedEncryptionKey(ctx context.Context, customObject v1alpha1.AWSConfig) error {
	ctlCtx, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		e.logger.LogCtx(ctx, "level", "debug", "message", "finding out encryption key")

		_, err := e.describeKey(ctx, customObject)
		if IsKeyNotFound(err) {
			e.logger.LogCtx(ctx, "level", "debug", "message", "did not find encryption key")

		} else if err != nil {
			return microerror.Mask(err)

		} else {

			e.logger.LogCtx(ctx, "level", "debug", "message", "found encryption key")
			e.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}
	}

	var keyID *string
	{
		e.logger.LogCtx(ctx, "level", "debug", "message", "creating encryption key")

		tags := key.ClusterTags(customObject, e.installationName)

		// TODO already created case should be handled here. Otherwise there is a chance alias wasn't created yet and EncryptionKey will tell there is no key. We can check tags to see if the key was created. Issue: https://github.com/giantswarm/giantswarm/issues/4262.
		in := &kms.CreateKeyInput{
			Tags: awstags.NewKMS(tags),
		}

		out, err := ctlCtx.AWSClient.KMS.CreateKey(in)
		if err != nil {
			return microerror.Mask(err)
		}

		keyID = out.KeyMetadata.KeyId

		e.logger.LogCtx(ctx, "level", "debug", "message", "created encryption key")
	}

	// This is importand key roation is enabled before creation alias.
	// Otherwise it is not guaranteed it will be reconciled. Right now we
	// *always* enable rotation before aliasing the key. So it is enough
	// reconcile aliasing properly to make sure rotation is enabled.
	{
		e.logger.LogCtx(ctx, "level", "debug", "message", "enabling encryption key rotation")

		in := &kms.EnableKeyRotationInput{
			KeyId: keyID,
		}

		_, err = ctlCtx.AWSClient.KMS.EnableKeyRotation(in)
		if err != nil {
			return microerror.Mask(err)
		}

		e.logger.LogCtx(ctx, "level", "debug", "message", "enabled encryption key rotation")
	}

	{
		e.logger.LogCtx(ctx, "level", "debug", "message", "creating encryption key alias")

		clusterID := key.ClusterID(customObject)
		keyAlias := aws.String(toAlias(clusterID))

		in := &kms.CreateAliasInput{
			AliasName:   keyAlias,
			TargetKeyId: keyID,
		}

		_, err = ctlCtx.AWSClient.KMS.CreateAlias(in)
		if err != nil {
			return microerror.Mask(err)
		}

		e.logger.LogCtx(ctx, "level", "debug", "message", "created encryption key alias")
	}

	return nil
}

func (e *Encrypter) EnsureDeletedEncryptionKey(ctx context.Context, customObject v1alpha1.AWSConfig) error {
	ctlCtx, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var keyID *string
	{
		e.logger.LogCtx(ctx, "level", "debug", "message", "finding out encryption key")

		// TODO we should search by tags here in case alias failed to create and cluster was deleted early. Issue: https://github.com/giantswarm/giantswarm/issues/4262.
		out, err := e.describeKey(ctx, customObject)
		if IsKeyNotFound(err) {
			e.logger.LogCtx(ctx, "level", "debug", "message", "did not find encryption key")
			e.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil

		} else if err != nil {
			return microerror.Mask(err)

		} else if out.KeyMetadata.DeletionDate != nil {
			e.logger.LogCtx(ctx, "level", "debug", "message", "encryption key is scheduled for deletion")
			e.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil

		} else {

			e.logger.LogCtx(ctx, "level", "debug", "message", "found encryption key")
			keyID = out.KeyMetadata.KeyId
		}
	}

	{
		e.logger.LogCtx(ctx, "level", "info", "message", "scheduling deletion of encryption key")

		// AWS API doesn't allow to delete the KMS key immediately, but
		// we can schedule its deletion. This also removes associated
		// aliases. 7 days is the smallest possible pending window.
		//
		// https://docs.aws.amazon.com/kms/latest/developerguide/deleting-keys.html

		pendingWindowInDays := aws.Int64(7)

		in := &kms.ScheduleKeyDeletionInput{
			KeyId:               keyID,
			PendingWindowInDays: pendingWindowInDays,
		}

		_, err = ctlCtx.AWSClient.KMS.ScheduleKeyDeletion(in)
		if err != nil {
			return microerror.Mask(err)
		}

		e.logger.LogCtx(ctx, "level", "info", "message", "scheduled deletion of encryption key")
	}

	return nil
}

func (k *Encrypter) CurrentState(ctx context.Context, customObject v1alpha1.AWSConfig) (encrypter.EncryptionKeyState, error) {
	var currentState encrypter.EncryptionKeyState

	clusterID := key.ClusterID(customObject)
	alias := toAlias(clusterID)

	output, err := k.describeKey(ctx, customObject)
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

func (k *Encrypter) EncryptionKey(ctx context.Context, customObject v1alpha1.AWSConfig) (string, error) {
	output, err := k.describeKey(ctx, customObject)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return *output.KeyMetadata.Arn, nil
}

func (k *Encrypter) Encrypt(ctx context.Context, key, plaintext string) (string, error) {
	sc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}

	encryptInput := &kms.EncryptInput{
		KeyId:     aws.String(key),
		Plaintext: []byte(plaintext),
	}

	encryptOutput, err := sc.AWSClient.KMS.Encrypt(encryptInput)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return string(encryptOutput.CiphertextBlob), nil
}

func (e *Encrypter) IsKeyNotFound(err error) bool {
	return IsKeyNotFound(err)
}

func (k *Encrypter) describeKey(ctx context.Context, customObject v1alpha1.AWSConfig) (*kms.DescribeKeyOutput, error) {
	sc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterID := key.ClusterID(customObject)
	alias := toAlias(clusterID)
	input := &kms.DescribeKeyInput{
		KeyId: aws.String(alias),
	}

	output, err := sc.AWSClient.KMS.DescribeKey(input)
	if IsKeyNotFound(err) {
		return nil, microerror.Mask(keyNotFoundError)
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	return output, nil
}

func toAlias(keyID string) string {
	return fmt.Sprintf("alias/%s", keyID)
}
