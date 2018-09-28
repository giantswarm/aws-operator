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
	Encrypter
	TLSManager
	Resource
}

type Encrypter interface {
	EncryptionKey(ctx context.Context, customObject v1alpha1.AWSConfig) (string, error)
	Encrypt(ctx context.Context, key, plaintext string) (string, error)
	IsKeyNotFound(error) bool
}

type Resource interface {
	EnsureCreatedEncryptionKey(context.Context, v1alpha1.AWSConfig) error
	EnsureDeletedEncryptionKey(context.Context, v1alpha1.AWSConfig) error
}

type RoleManager interface {
	EnsureCreatedAuthorizedIAMRoles(context.Context, v1alpha1.AWSConfig) error
	EnsureDeletedAuthorizedIAMRoles(context.Context, v1alpha1.AWSConfig) error
}

type TLSManager interface {
	EncryptTLSAssets(ctx context.Context, customObject v1alpha1.AWSConfig, assets legacy.AssetsBundle) (*legacy.CompactTLSAssets, error)
}
