package healthz

// Response is the return value of the service action.
type Response struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// DefaultResponse provides a default response object by best effort.
func DefaultResponse() *Response {
	return &Response{
		Code:    "",
		Message: "",
	}
}
