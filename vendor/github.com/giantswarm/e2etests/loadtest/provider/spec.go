package provider

import "context"

type Interface interface {
	InstallTestApp(ctx context.Context) error
	WaitForTestApp(ctx context.Context) error
}
