package encrypter

import (
	"context"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha3"
)

type Mock struct {
	IsError bool
	KeyID   string
	KeyName string
}

func (m *Mock) Encrypt(ctx context.Context, key, plaintext string) (string, error) {
	return plaintext, nil
}

func (m *Mock) EncryptionKey(ctx context.Context, id string) (string, error) {
	return "", nil
}

func (m *Mock) EnsureCreatedEncryptionKey(context.Context, infrastructurev1alpha3.AWSCluster) error {
	return nil
}

func (m *Mock) EnsureDeletedEncryptionKey(context.Context, infrastructurev1alpha3.AWSCluster) error {
	return nil
}

func (m *Mock) IsKeyNotFound(err error) bool {
	return false
}
