package encrypter

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs/legacy"
)

const (
	KMSBackend   = "kms"
	VaultBackend = "vault"
)

type EncryptionKeyState struct {
	KeyID   string
	KeyName string
}

type Interface interface {
	KeyManager
	StateManager
	TLSManager
	Encrypter
}

type KeyManager interface {
	CreateKey(ctx context.Context, customObject v1alpha1.AWSConfig, keyName string) error
	DeleteKey(ctx context.Context, keyName string) error
}

type StateManager interface {
	CurrentState(ctx context.Context, customObject v1alpha1.AWSConfig) (EncryptionKeyState, error)
	DesiredState(ctx context.Context, customObject v1alpha1.AWSConfig) (EncryptionKeyState, error)
}

type TLSManager interface {
	EncryptTLSAssets(ctx context.Context, customObject v1alpha1.AWSConfig, assets legacy.AssetsBundle) (*legacy.CompactTLSAssets, error)
}

type Encrypter interface {
	EncryptionKey(ctx context.Context, customObject v1alpha1.AWSConfig) (string, error)
	Encrypt(ctx context.Context, key, plaintext string) (string, error)
}
