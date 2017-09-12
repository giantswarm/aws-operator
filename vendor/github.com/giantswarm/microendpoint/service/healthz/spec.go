package healthz

import "context"

// Response is the response structure each health check returns.
type Response struct {
	Description string `json:"description" yaml:"description"`
	Failed      bool   `json:"failed" yaml:"failed"`
	Message     string `json:"message" yaml:"message"`
	Name        string `json:"name" yaml:"name"`
}

type Responses []Response

func (rs Responses) HasFailed() bool {
	for _, r := range rs {
		if r.Failed {
			return true
		}
	}

	return false
}

// Service describes how an implementation of a health check has to be done.
type Service interface {
	// GetHealthz executes the health check business logic. It always returns a
	// response structure containing information about the system state it checks.
	// GetHealthz only returns an error in case it cannot create its response
	// structure properly. This implies that failed health checks must not be
	// represented by an error being returned, but by the response structure
	// itself.
	GetHealthz(ctx context.Context) (Response, error)
}
