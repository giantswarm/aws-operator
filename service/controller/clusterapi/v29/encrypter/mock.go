package encrypter

import (
	"context"

	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

type EncrypterMock struct {
	IsError bool
	KeyID   string
	KeyName string
}

func (e *EncrypterMock) Encrypt(ctx context.Context, key, plaintext string) (string, error) {
	return plaintext, nil
}

func (e *EncrypterMock) EncryptionKey(ctx context.Context, customObject v1alpha1.Cluster) (string, error) {
	return "", nil
}

func (e *EncrypterMock) EnsureCreatedEncryptionKey(context.Context, v1alpha1.Cluster) error {
	return nil
}

func (e *EncrypterMock) EnsureDeletedEncryptionKey(context.Context, v1alpha1.Cluster) error {
	return nil
}

func (e *EncrypterMock) IsKeyNotFound(err error) bool {
	return false
}
