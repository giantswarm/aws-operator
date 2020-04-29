package encrypter

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
)

type EncrypterMock struct {
	IsError bool
	KeyID   string
	KeyName string
}

func (e *EncrypterMock) Encrypt(ctx context.Context, key, plaintext string) (string, error) {
	ciphertext := base64.StdEncoding.EncodeToString([]byte(plaintext))
	return ciphertext, nil
}

func (e *EncrypterMock) Decrypt(ctx context.Context, key, ciphertext string) (string, error) {
	plaintext, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(ciphertext, "data:text/plain;charset=utf-8;base64,"))
	if err != nil {
		return "", microerror.Mask(err)
	}
	return string(plaintext), nil
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
