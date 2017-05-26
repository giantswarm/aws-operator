package aws

import (
	"github.com/giantswarm/awstpr/aws/hostedzones"
	"github.com/giantswarm/awstpr/aws/vpc"
)

type AWS struct {
	Masters     []Node                  `json:"masters" yaml:"masters"`
	Workers     []Node                  `json:"workers" yaml:"workers"`
	Region      string                  `json:"region" yaml:"region"`
	AZ          string                  `json:"az" yaml:"az"`
	HostedZones hostedzones.HostedZones `json:"hostedZones" yaml:"hostedZones"`
	VPC         vpc.VPC                 `json:"vpc" yaml:"vpc"`
}
