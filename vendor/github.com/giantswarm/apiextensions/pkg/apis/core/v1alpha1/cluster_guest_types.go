package v1alpha1

type ClusterGuestConfig struct {
	API            ClusterGuestConfigAPI             `json:"api" yaml:"api"`
	ID             string                            `json:"id" yaml:"id"`
	Name           string                            `json:"name,omitempty" yaml:"name,omitempty"`
	Owner          string                            `json:"owner,omitempty" yaml:"owner,omitempty"`
	VersionBundles []ClusterGuestConfigVersionBundle `json:"versionBundles,omitempty" yaml:"versionBundles,omitempty"`
}

type ClusterGuestConfigAPI struct {
	Endpoint string `json:"endpoint" yaml:"endpoint"`
}

type ClusterGuestConfigVersionBundle struct {
	Name    string `json:"name" yaml:"name"`
	Version string `json:"version" yaml:"version"`
}
