package spec

import (
	"github.com/giantswarm/awstpr/spec/aws"
)

type AWS struct {
	Masters     []aws.Node      `json:"masters" yaml:"masters"`
	Workers     []aws.Node      `json:"workers" yaml:"workers"`
	Region      string          `json:"region" yaml:"region"`
	AZ          string          `json:"az" yaml:"az"`
	ELB         aws.ELB         `json:"elb" yaml:"elb"`
	HostedZones aws.HostedZones `json:"hostedZones" yaml:"hostedZones"`
	VPC         aws.VPC         `json:"vpc" yaml:"vpc"`
}
