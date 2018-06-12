package encrypter

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

type EncryptionKeyState struct {
	KeyID   string
	KeyName string
}

type Interface interface {
	CreateKey(ctx context.Context, customObject v1alpha1.AWSConfig, keyName string) error
	DeleteKey(ctx context.Context, keyName string) error
	CurrentState(ctx context.Context, customObject v1alpha1.AWSConfig) (EncryptionKeyState, error)
	DesiredState(ctx context.Context, customObject v1alpha1.AWSConfig) (EncryptionKeyState, error)
}
