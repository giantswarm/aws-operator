package encrypter

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

type EncrypterMock struct {
	IsError bool
	KeyID   string
	KeyName string
}

func (e *EncrypterMock) Encrypt(ctx context.Context, key, plaintext string) (string, error) {
	return plaintext, nil
}

func (e *EncrypterMock) Decrypt(ctx context.Context, key, ciphertext string) (string, error) {
	return ciphertext, nil
}

func (e *EncrypterMock) EncryptionKey(ctx context.Context, customObject v1alpha1.AWSConfig) (string, error) {
	return "", nil
}

func (e *EncrypterMock) EnsureCreatedEncryptionKey(context.Context, v1alpha1.AWSConfig) error {
	return nil
}

func (e *EncrypterMock) EnsureDeletedEncryptionKey(context.Context, v1alpha1.AWSConfig) error {
	return nil
}

func (e *EncrypterMock) IsKeyNotFound(err error) bool {
	return false
}
