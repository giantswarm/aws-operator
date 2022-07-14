package kms

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/v12/pkg/awstags"
	"github.com/giantswarm/aws-operator/v12/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v12/service/controller/key"
)

type EncrypterConfig struct {
	Logger micrologger.Logger

	InstallationName string
}

type Encrypter struct {
	logger micrologger.Logger

	cache *Cache

	installationName string
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

		cache: NewCache(),

		installationName: c.InstallationName,
	}

	return kms, nil
}

func (e *Encrypter) EnsureCreatedEncryptionKey(ctx context.Context, cr infrastructurev1alpha3.AWSCluster) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var oldKeyScheduledForDeletion bool
	{
		e.logger.Debugf(ctx, "finding encryption key")

		_, err := e.cachedKey(ctx, key.ClusterID(&cr))
		if IsKeyNotFound(err) {
			e.logger.Debugf(ctx, "did not find encryption key")

		} else if IsKeyScheduledForDeletion(err) {
			e.logger.Debugf(ctx, "found encryption key")
			e.logger.Debugf(ctx, "current encryption key is scheduled for deletion and will be recreated")
			oldKeyScheduledForDeletion = true

		} else if err != nil {
			return microerror.Mask(err)

		} else {
			e.logger.Debugf(ctx, "found encryption key")
			e.logger.Debugf(ctx, "canceling resource")
			return nil
		}
	}

	if oldKeyScheduledForDeletion {
		// In such case we just delete the alias so we can alias newly
		// created key.

		e.logger.Debugf(ctx, "deleting old encryption key alias")

		in := &kms.DeleteAliasInput{
			AliasName: aws.String(keyAlias(key.ClusterID(&cr))),
		}

		_, err = cc.Client.TenantCluster.AWS.KMS.DeleteAlias(in)
		if err != nil {
			return microerror.Mask(err)
		}

		e.logger.Debugf(ctx, "deleted old encryption key alias")
	}

	var keyID *string
	{
		e.logger.Debugf(ctx, "creating encryption key")

		tags := key.AWSTags(&cr, e.installationName)

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

		e.logger.Debugf(ctx, "created encryption key")
	}

	// NOTE: Key roation must be enabled before creation alias. Otherwise
	// it is not guaranteed it will be reconciled. Right now we *always*
	// enable rotation before aliasing the key. So it is enough reconcile
	// aliasing properly to make sure rotation is enabled.
	{
		e.logger.Debugf(ctx, "enabling encryption key rotation")

		in := &kms.EnableKeyRotationInput{
			KeyId: keyID,
		}

		_, err = cc.Client.TenantCluster.AWS.KMS.EnableKeyRotation(in)
		if err != nil {
			return microerror.Mask(err)
		}

		e.logger.Debugf(ctx, "enabled encryption key rotation")
	}

	{
		e.logger.Debugf(ctx, "creating encryption key alias")

		in := &kms.CreateAliasInput{
			AliasName:   aws.String(keyAlias(key.ClusterID(&cr))),
			TargetKeyId: keyID,
		}

		_, err = cc.Client.TenantCluster.AWS.KMS.CreateAlias(in)
		if err != nil {
			return microerror.Mask(err)
		}

		e.logger.Debugf(ctx, "created encryption key alias")
	}

	return nil
}

func (e *Encrypter) EnsureDeletedEncryptionKey(ctx context.Context, cr infrastructurev1alpha3.AWSCluster) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var keyID *string
	{
		e.logger.Debugf(ctx, "finding encryption key")

		// TODO we should search by tags here in case alias failed to create and
		// cluster was deleted early.
		//
		//     https://github.com/giantswarm/giantswarm/issues/4262.
		//
		out, err := e.cachedKey(ctx, key.ClusterID(&cr))
		if IsKeyNotFound(err) || IsKeyScheduledForDeletion(err) {
			e.logger.Debugf(ctx, "did not find encryption key")
			e.logger.Debugf(ctx, "canceling resource")
			return nil

		} else if err != nil {
			return microerror.Mask(err)

		} else if out.KeyMetadata.DeletionDate != nil {
			e.logger.Debugf(ctx, "encryption key is scheduled for deletion")
			e.logger.Debugf(ctx, "canceling resource")
			return nil

		} else {

			e.logger.Debugf(ctx, "found encryption key")
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

func (e *Encrypter) EncryptionKey(ctx context.Context, id string) (string, error) {
	out, err := e.cachedKey(ctx, id)
	if err != nil {
		return "", microerror.Mask(err)
	}

	// When key is scheduled for deletion we consider it deleted.
	if out.KeyMetadata.DeletionDate != nil {
		return "", microerror.Mask(keyNotFoundError)
	}

	return *out.KeyMetadata.Arn, nil
}

func (e *Encrypter) Encrypt(ctx context.Context, key, plaintext string) (string, error) {
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

func (e *Encrypter) cachedKey(ctx context.Context, id string) (*kms.DescribeKeyOutput, error) {
	var err error
	var ok bool

	var keyOutput *kms.DescribeKeyOutput
	{
		ck := e.cache.Key(ctx, id)

		if ck == "" {
			keyOutput, err = e.lookupKey(ctx, id)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		} else {
			keyOutput, ok = e.cache.Get(ctx, ck)
			if !ok {
				keyOutput, err = e.lookupKey(ctx, id)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				e.cache.Set(ctx, ck, keyOutput)
			}
		}
	}

	return keyOutput, nil
}

func (e *Encrypter) lookupKey(ctx context.Context, id string) (*kms.DescribeKeyOutput, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	input := &kms.DescribeKeyInput{
		KeyId: aws.String(keyAlias(id)),
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
