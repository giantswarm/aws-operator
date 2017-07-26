package healthz

// Request is the configuration for the service action.
type Request struct {
}

// DefaultRequest provides a default request object by best effort.
func DefaultRequest() Request {
	return Request{}
}
