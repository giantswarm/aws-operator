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
	// DeleteRelease removes a release from the server.
	DeleteRelease(ctx context.Context, name, release string) error
	// GetReleaseVersion queries CNR for the release version of the chart
	// represented by the given name and channel.
	GetReleaseVersion(ctx context.Context, name, channel string) (string, error)
	// PromoteChart puts a release of the given chart in a channel.
	PromoteChart(ctx context.Context, name, release, channel string) error
	// PullChartTarball downloads a tarball with the chart described by
	// the given chart name and channel, returning the file path.
	PullChartTarball(ctx context.Context, name, channel string) (string, error)
	// PullChartTarballFromRelease downloads a tarball with the chart described
	// by the given chart name and release, returning the file path.
	PullChartTarballFromRelease(ctx context.Context, name, release string) (string, error)
	// PushChartTarball sends a tarball to the server to be installed for the given
	// name and release.
	PushChartTarball(ctx context.Context, name, release, tarballPath string) error
}
