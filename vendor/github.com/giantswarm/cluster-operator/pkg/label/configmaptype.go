package label

const (
	// ConfigMapType is a type of configmap used for tenant clusters.
	ConfigMapType = "cluster-operator.giantswarm.io/configmap-type"
)

const (
	// ConfigMapTypeApp is a label value for app configmaps managed by the
	// operator.
	ConfigMapTypeApp = "app"
	// ConfigMapTypeUser is a label value for user configmaps created by the
	// operator and edited by users to override chart values.
	ConfigMapTypeUser = "user"
)
