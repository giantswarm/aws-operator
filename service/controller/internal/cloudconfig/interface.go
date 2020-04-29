package cloudconfig

import (
	"context"
)

type Interface interface {
	DecryptTemplate(ctx context.Context, data string) (string, error)

	NewMasterTemplate(ctx context.Context, data IgnitionTemplateData) (string, error)
	NewWorkerTemplate(ctx context.Context, data IgnitionTemplateData) (string, error)
}
