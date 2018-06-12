package encrypter

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs/legacy"
)

type EncrypterMock struct {
	IsError bool
	KeyID   string
	KeyName string
}

func (e *EncrypterMock) CreateKey(ctx context.Context, customObject v1alpha1.AWSConfig, keyName string) error {
	return nil
}

func (e *EncrypterMock) DeleteKey(ctx context.Context, keyName string) error {
	return nil
}

func (e *EncrypterMock) CurrentState(ctx context.Context, customObject v1alpha1.AWSConfig) (EncryptionKeyState, error) {
	if e.IsError {
		return EncryptionKeyState{}, fmt.Errorf("could not get current state")
	}

	return EncryptionKeyState{
		KeyID:   e.KeyID,
		KeyName: e.KeyName,
	}, nil
}

func (e *EncrypterMock) DesiredState(ctx context.Context, customObject v1alpha1.AWSConfig) (EncryptionKeyState, error) {
	if e.IsError {
		return EncryptionKeyState{}, fmt.Errorf("could not get current state")
	}

	return EncryptionKeyState{
		KeyID:   e.KeyID,
		KeyName: e.KeyName,
	}, nil
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
