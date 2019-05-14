package encrypter

import (
	"context"

	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
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
	EncryptionKey(ctx context.Context, customObject v1alpha1.Cluster) (string, error)
	Encrypt(ctx context.Context, key, plaintext string) (string, error)
	IsKeyNotFound(error) bool
}

type Resource interface {
	EnsureCreatedEncryptionKey(context.Context, v1alpha1.Cluster) error
	EnsureDeletedEncryptionKey(context.Context, v1alpha1.Cluster) error
}

type RoleManager interface {
	EnsureCreatedAuthorizedIAMRoles(context.Context, v1alpha1.Cluster) error
	EnsureDeletedAuthorizedIAMRoles(context.Context, v1alpha1.Cluster) error
}
