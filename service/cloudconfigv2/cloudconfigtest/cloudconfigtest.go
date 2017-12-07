package cloudconfigtest

import (
	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-operator/service/cloudconfigv2"
)

func New() *cloudconfigv2.CloudConfig {
	c := cloudconfigv2.DefaultConfig()

	c.Logger = microloggertest.New()

	newCloudConfig, err := cloudconfigv2.New(c)
	if err != nil {
		panic(err)
	}

	return newCloudConfig
}
