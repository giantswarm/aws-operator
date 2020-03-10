package label

const (
	// ServiceType is a standard label for guest resources.
	ServiceType = "giantswarm.io/service-type"
)

const (
	// ServiceTypeManaged is a label value for managed resources.
	ServiceTypeManaged = "managed"
	// ServiceTypeSystem is a label value for system resources.
	ServiceTypeSystem = "system"
)
