package aws

import "github.com/giantswarm/awstpr/spec/aws/elb"

// ELB contains additional settings for the elastic load balancers.
type ELB struct {
	// IdleTimeoutSeconds is idle time before closing the front-end and back-end connections
	IdleTimeoutSeconds elb.IdleTimeoutSeconds `json:"idleTimeoutSeconds" yaml:"idleTimeoutSeconds"`
}
