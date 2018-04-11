package apprclient

const (
	httpClientTimeout = 5
)

// Interface describes the methods provided by the appr client.
type Interface interface {
	GetReleaseVersion(name, channel string) (string, error)
	PullChartTarball(name, channel string) (string, error)
}

// Channel represents a CNR channel.
type Channel struct {
	Current string `json:"current"`
}
