package encrypter

import (
	"context"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
)

const (
	KMSBackend   = "kms"
	VaultBackend = "vault"
)

type Interface interface {
	Encrypter
	Resource
}

type Encrypter interface {
	EncryptionKey(ctx context.Context, customObject infrastructurev1alpha2.AWSCluster) (string, error)
	Encrypt(ctx context.Context, key, plaintext string) (string, error)
	IsKeyNotFound(error) bool
}

type Resource interface {
	EnsureCreatedEncryptionKey(context.Context, infrastructurev1alpha2.AWSCluster) error
	EnsureDeletedEncryptionKey(context.Context, infrastructurev1alpha2.AWSCluster) error
}

type RoleManager interface {
	EnsureCreatedAuthorizedIAMRoles(context.Context, infrastructurev1alpha2.AWSCluster) error
	EnsureDeletedAuthorizedIAMRoles(context.Context, infrastructurev1alpha2.AWSCluster) error
}
