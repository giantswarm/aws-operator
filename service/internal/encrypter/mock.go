package encrypter

import (
	"context"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
)

type EncrypterMock struct {
	IsError bool
	KeyID   string
	KeyName string
}

func (e *EncrypterMock) Encrypt(ctx context.Context, key, plaintext string) (string, error) {
	return plaintext, nil
}

func (e *EncrypterMock) EncryptionKey(ctx context.Context, customObject infrastructurev1alpha2.AWSCluster) (string, error) {
	return "", nil
}

func (e *EncrypterMock) EnsureCreatedEncryptionKey(context.Context, infrastructurev1alpha2.AWSCluster) error {
	return nil
}

func (e *EncrypterMock) EnsureDeletedEncryptionKey(context.Context, infrastructurev1alpha2.AWSCluster) error {
	return nil
}

func (e *EncrypterMock) IsKeyNotFound(err error) bool {
	return false
}
