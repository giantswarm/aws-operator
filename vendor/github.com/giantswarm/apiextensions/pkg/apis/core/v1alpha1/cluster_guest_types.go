package v1alpha1

type ClusterGuestConfig struct {
	AvailabilityZones int `json:"availabilityZones,omitempty" yaml:"availabilityZones,omitempty"`
	// DNSZone for guest cluster is supplemented with host prefixes for
	// specific services such as Kubernetes API or Etcd. In general this DNS
	// Zone should start with `k8s` like for example
	// `k8s.cluster.example.com.`.
	DNSZone        string                            `json:"dnsZone" yaml:"dnsZone"`
	ID             string                            `json:"id" yaml:"id"`
	Name           string                            `json:"name,omitempty" yaml:"name,omitempty"`
	Owner          string                            `json:"owner,omitempty" yaml:"owner,omitempty"`
	ReleaseVersion string                            `json:"releaseVersion,omitempty" yaml:"releaseVersion,omitempty"`
	Scaling        ClusterScaling                    `json:"scaling" yaml:"scaling"`
	VersionBundles []ClusterGuestConfigVersionBundle `json:"versionBundles,omitempty" yaml:"versionBundles,omitempty"`
}

type ClusterGuestConfigVersionBundle struct {
	Name    string `json:"name" yaml:"name"`
	Version string `json:"version" yaml:"version"`
}

type ClusterScaling struct {
	// Max defines maximum number of worker nodes guest cluster is allowed to have.
	Max int `json:"max" yaml:"max"`
	// Min defines minimum number of worker nodes required to be present in guest cluster.
	Min int `json:"min" yaml:"min"`
}
