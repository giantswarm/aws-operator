package encrypter

import (
	"context"
	"fmt"
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
	ciphertext := fmt.Sprintf("<encrypted>--%s", plaintext)
	return ciphertext, nil
}

func (e *EncrypterMock) Decrypt(ctx context.Context, key, ciphertext string) (string, error) {
	if !strings.HasPrefix(ciphertext, "<encrypted>--") {
		return "", microerror.Mask(fmt.Errorf("InvalidCiphertextException"))
	}
	plaintext := strings.TrimPrefix(ciphertext, "<encrypted>--")
	return plaintext, nil
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
