// Package messagecontext stores and accesses the message struct in
// context.Context.
package messagecontext

import (
	"context"
)

// key is an unexported type for keys defined in this package. This prevents
// collisions with keys defined in other packages.
type key string

// messageKey is the key for message struct values in context.Context. Clients
// use messagecontext.NewContext and messagecontext.FromContext instead of using
// this key directly.
var messageKey key = "message"

// Message is a communication structure used to transport information from one
// resource to another. Messages move between resource during reconciliation
// within the dispatched context.
type Message struct {
	// ConfigMapNames is a list of config map names filled by the config map
	// resource and read by the deployment resource.
	ConfigMapNames []string
}

// NewMessage returns a new communication structure used to apply to a context.
func NewMessage() *Message {
	return &Message{}
}

// NewContext returns a new context.Context that carries value v.
func NewContext(ctx context.Context, v *Message) context.Context {
	if v == nil {
		return ctx
	}

	return context.WithValue(ctx, messageKey, v)
}

// FromContext returns the message struct, if any.
func FromContext(ctx context.Context) (*Message, bool) {
	v, ok := ctx.Value(messageKey).(*Message)
	return v, ok
}
