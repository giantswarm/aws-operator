package cloudconfigtest

import (
	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-operator/service/cloudconfigv3"
)

func New() *cloudconfigv3.CloudConfig {
	c := cloudconfigv3.DefaultConfig()

	c.Logger = microloggertest.New()

	newCloudConfig, err := cloudconfigv3.New(c)
	if err != nil {
		panic(err)
	}

	return newCloudConfig
}
