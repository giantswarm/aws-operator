package ssm

import (
	"github.com/giantswarm/awstpr/aws/ssm/docker"
)

type SSMAgent struct {
	Docker docker.Docker `json:"docker" yaml:"docker"`
}
