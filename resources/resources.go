package resources

type Resource interface {
	CreateIfNotExists() (bool, error)
	CreateOrFail() error
	Delete() error
}

type ResourceWithID interface {
	ID() string
	Resource
}

type NamedResource interface {
	Name() string
	Resource
}

type ArnResource interface {
	Arn() string
	Resource
}
