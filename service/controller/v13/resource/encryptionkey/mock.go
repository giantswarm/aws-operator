package encryptionkey

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/v13/encrypter"
)

type EncrypterMock struct {
	isError bool
	keyID   string
	keyName string
}

func (e *EncrypterMock) CreateKey(ctx context.Context, customObject v1alpha1.AWSConfig, keyName string) error {
	return nil
}

func (e *EncrypterMock) DeleteKey(ctx context.Context, keyName string) error {
	return nil
}

func (e *EncrypterMock) CurrentState(ctx context.Context, customObject v1alpha1.AWSConfig) (encrypter.EncryptionKeyState, error) {
	if e.isError {
		return encrypter.EncryptionKeyState{}, fmt.Errorf("could not get current state")
	}

	return encrypter.EncryptionKeyState{
		KeyID:   e.keyID,
		KeyName: e.keyName,
	}, nil
}

func (e *EncrypterMock) DesiredState(ctx context.Context, customObject v1alpha1.AWSConfig) (encrypter.EncryptionKeyState, error) {
	if e.isError {
		return encrypter.EncryptionKeyState{}, fmt.Errorf("could not get current state")
	}

	return encrypter.EncryptionKeyState{
		KeyID:   e.keyID,
		KeyName: e.keyName,
	}, nil
}
