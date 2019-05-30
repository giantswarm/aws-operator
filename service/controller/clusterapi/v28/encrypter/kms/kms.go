package kms

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/pkg/awstags"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/key"
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

func (e *Encrypter) EnsureCreatedEncryptionKey(ctx context.Context, cr v1alpha1.Cluster) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var oldKeyScheduledForDeletion bool
	{
		e.logger.LogCtx(ctx, "level", "debug", "message", "finding out encryption key")

		_, err := e.describeKey(ctx, cr)
		if IsKeyNotFound(err) {
			e.logger.LogCtx(ctx, "level", "debug", "message", "did not find encryption key")

		} else if IsKeyScheduledForDeletion(err) {
			e.logger.LogCtx(ctx, "level", "debug", "message", "found encryption key")
			e.logger.LogCtx(ctx, "level", "debug", "message", "current encryption key is scheduled for deletion and will be recreated")
			oldKeyScheduledForDeletion = true

		} else if err != nil {
			return microerror.Mask(err)

		} else {
			e.logger.LogCtx(ctx, "level", "debug", "message", "found encryption key")
			e.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}
	}

	if oldKeyScheduledForDeletion {
		// In such case we just delete the alias so we can alias newly
		// created key.

		e.logger.LogCtx(ctx, "level", "debug", "message", "deleting old encryption key alias")

		in := &kms.DeleteAliasInput{
			AliasName: aws.String(keyAlias(cr)),
		}

		_, err = cc.Client.TenantCluster.AWS.KMS.DeleteAlias(in)
		if err != nil {
			return microerror.Mask(err)
		}

		e.logger.LogCtx(ctx, "level", "debug", "message", "deleted old encryption key alias")
	}

	var keyID *string
	{
		e.logger.LogCtx(ctx, "level", "debug", "message", "creating encryption key")

		tags := key.ClusterTags(cr, e.installationName)

		// TODO already created case should be handled here. Otherwise there is a
		// chance alias wasn't created yet and EncryptionKey will tell there is no
		// key. We can check tags to see if the key was created.
		//
		//     https://github.com/giantswarm/giantswarm/issues/4262
		//
		in := &kms.CreateKeyInput{
			Tags: awstags.NewKMS(tags),
		}

		out, err := cc.Client.TenantCluster.AWS.KMS.CreateKey(in)
		if err != nil {
			return microerror.Mask(err)
		}

		keyID = out.KeyMetadata.KeyId

		e.logger.LogCtx(ctx, "level", "debug", "message", "created encryption key")
	}

	// NOTE: Key roation must be enabled before creation alias. Otherwise
	// it is not guaranteed it will be reconciled. Right now we *always*
	// enable rotation before aliasing the key. So it is enough reconcile
	// aliasing properly to make sure rotation is enabled.
	{
		e.logger.LogCtx(ctx, "level", "debug", "message", "enabling encryption key rotation")

		in := &kms.EnableKeyRotationInput{
			KeyId: keyID,
		}

		_, err = cc.Client.TenantCluster.AWS.KMS.EnableKeyRotation(in)
		if err != nil {
			return microerror.Mask(err)
		}

		e.logger.LogCtx(ctx, "level", "debug", "message", "enabled encryption key rotation")
	}

	{
		e.logger.LogCtx(ctx, "level", "debug", "message", "creating encryption key alias")

		in := &kms.CreateAliasInput{
			AliasName:   aws.String(keyAlias(cr)),
			TargetKeyId: keyID,
		}

		_, err = cc.Client.TenantCluster.AWS.KMS.CreateAlias(in)
		if err != nil {
			return microerror.Mask(err)
		}

		e.logger.LogCtx(ctx, "level", "debug", "message", "created encryption key alias")
	}

	return nil
}

func (e *Encrypter) EnsureDeletedEncryptionKey(ctx context.Context, cr v1alpha1.Cluster) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var keyID *string
	{
		e.logger.LogCtx(ctx, "level", "debug", "message", "finding out encryption key")

		// TODO we should search by tags here in case alias failed to create and cluster was deleted early. Issue: https://github.com/giantswarm/giantswarm/issues/4262.
		out, err := e.describeKey(ctx, cr)
		if IsKeyNotFound(err) || IsKeyScheduledForDeletion(err) {
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

		_, err = cc.Client.TenantCluster.AWS.KMS.ScheduleKeyDeletion(in)
		if err != nil {
			return microerror.Mask(err)
		}

		e.logger.LogCtx(ctx, "level", "info", "message", "scheduled deletion of encryption key")
	}

	return nil
}

func (k *Encrypter) EncryptionKey(ctx context.Context, cr v1alpha1.Cluster) (string, error) {
	out, err := k.describeKey(ctx, cr)
	if err != nil {
		return "", microerror.Mask(err)
	}

	// When key is scheduled for deletion we consider it deleted.
	if out.KeyMetadata.DeletionDate != nil {
		return "", microerror.Mask(keyNotFoundError)
	}

	return *out.KeyMetadata.Arn, nil
}

func (k *Encrypter) Encrypt(ctx context.Context, key, plaintext string) (string, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}

	encryptInput := &kms.EncryptInput{
		KeyId:     aws.String(key),
		Plaintext: []byte(plaintext),
	}

	encryptOutput, err := cc.Client.TenantCluster.AWS.KMS.Encrypt(encryptInput)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return string(encryptOutput.CiphertextBlob), nil
}

func (e *Encrypter) IsKeyNotFound(err error) bool {
	return IsKeyNotFound(err) || IsKeyScheduledForDeletion(err)
}

func (k *Encrypter) describeKey(ctx context.Context, cr v1alpha1.Cluster) (*kms.DescribeKeyOutput, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	input := &kms.DescribeKeyInput{
		KeyId: aws.String(keyAlias(cr)),
	}

	out, err := cc.Client.TenantCluster.AWS.KMS.DescribeKey(input)
	if IsKeyNotFound(err) {
		return nil, microerror.Mask(keyNotFoundError)
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	if out.KeyMetadata.DeletionDate != nil {
		return nil, microerror.Mask(keyScheduledForDeletionError)
	}

	return out, nil
}
