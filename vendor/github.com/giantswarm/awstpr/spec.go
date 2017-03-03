package awstpr

import (
	"github.com/giantswarm/awstpr/aws"
	"github.com/giantswarm/clustertpr"
)

type Spec struct {
	Cluster clustertpr.Cluster `json:"cluster" yaml:"cluster"`
	AWS     aws.AWS            `json:"aws" yaml:"aws"`
}
