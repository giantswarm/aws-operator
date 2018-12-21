package apprclient

import "context"

const (
	httpClientTimeout = 5

	// okStauts is the status returned on a successful GET request to a CNR server.
	okStauts = "ok"
	// deletedStatus is the status returned on a successful DELETE request to a
	// CNR server.
	deletedStatus = "deleted"
)

// Interface describes the methods provided by the appr client.
type Interface interface {
	GetReleaseVersion(ctx context.Context, name, channel string) (string, error)
	PullChartTarball(ctx context.Context, name, channel string) (string, error)
	PullChartTarballFromRelease(ctx context.Context, name, release string) (string, error)
}
