package resources

type Resource interface {
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
