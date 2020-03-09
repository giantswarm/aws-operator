package key

// AppSpec is used to define app custom resources.
type AppSpec struct {
	App     string
	Catalog string
	Chart   string
	// Whether app is installed for clusterapi clusters only.
	ClusterAPIOnly bool
	// ConfigMapName overrides the name, otherwise the cluster values configmap
	// is used.
	ConfigMapName string
	// Whether app is installed for legacy clusters only.
	LegacyOnly      bool
	Namespace       string
	UseUpgradeForce bool
	Version         string
}
