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
	Encrypt(ctx context.Context, key, plaintext string) (string, error)
	EncryptTLSAssets(ctx context.Context, customObject v1alpha1.AWSConfig, assets legacy.AssetsBundle) (*legacy.CompactTLSAssets, error)
	EncryptionKey(ctx context.Context, customObject v1alpha1.AWSConfig) (string, error)
	IsKeyNotFound(error) bool
}

type Resource interface {
	CreateKey(ctx context.Context, customObject v1alpha1.AWSConfig, keyName string) error
	DeleteKey(ctx context.Context, keyName string) error
	GetCurrentState(ctx context.Context, customObject v1alpha1.AWSConfig) (EncryptionKeyState, error)
	GetDesiredState(ctx context.Context, customObject v1alpha1.AWSConfig) (EncryptionKeyState, error)

	EnsureCreatedAuthorizedIAMRoles(ctx context.Context, iamRoleARNs ...string) error
	EnsureDeletedAuthorizedIAMRoles(ctx context.Context, iamRoleARNs ...string) error
}
