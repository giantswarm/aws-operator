package resources

type Resource interface {
	// CreateIfNotExists creates a resource, unless it was already there, in which case it reuses it
	// the first return value is false when the resource has been reused, true when it has been created
	CreateIfNotExists() (bool, error)
	CreateOrFail() error
	Delete() error
}

type NamedResource interface {
	Name() string
	Resource
}

type ArnResource interface {
	Arn() string
	Resource
}

type ResourceWithID interface {
	ID() string
	Resource
}

type DNSNamedResource interface {
	DNSName() string
	HostedZoneID() string
	Resource
}
