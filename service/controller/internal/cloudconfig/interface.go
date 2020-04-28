package cloudconfig

import (
	"context"
)

type Interface interface {
	NewMasterTemplate(ctx context.Context, data IgnitionTemplateData) (string, string, error)
	NewWorkerTemplate(ctx context.Context, data IgnitionTemplateData) (string, string, error)
	DecryptedHash(ctx context.Context, data []byte) (string, error)
}
