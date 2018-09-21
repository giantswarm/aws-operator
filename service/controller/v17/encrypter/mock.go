package encrypter

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs/legacy"
)

type EncrypterMock struct {
	IsError bool
	KeyID   string
	KeyName string
}

func (e *EncrypterMock) EncryptionKey(ctx context.Context, customObject v1alpha1.AWSConfig) (string, error) {
	return "", nil
}

func (e *EncrypterMock) Encrypt(ctx context.Context, key, plaintext string) (string, error) {
	return plaintext, nil
}

func (e *EncrypterMock) EncryptTLSAssets(ctx context.Context, customObject v1alpha1.AWSConfig, assets legacy.AssetsBundle) (*legacy.CompactTLSAssets, error) {
	return &legacy.CompactTLSAssets{}, nil
}

func (e *EncrypterMock) IsKeyNotFound(err error) bool {
	return false
}
