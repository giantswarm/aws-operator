package encrypter

import (
	"context"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
)

const (
	KMSBackend = "kms"
)

type Interface interface {
	Encrypter
	Resource
}

type Encrypter interface {
	// EncryptionKey fetches the KMS encryption key for the Tenant Cluster
	// defined by id.
	EncryptionKey(ctx context.Context, id string) (string, error)
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
