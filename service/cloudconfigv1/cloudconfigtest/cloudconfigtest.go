package cloudconfigtest

import (
	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-operator/service/cloudconfigv1"
)

func New() *cloudconfigv1.CloudConfig {
	c := cloudconfigv1.DefaultConfig()

	c.Logger = microloggertest.New()

	newCloudConfig, err := cloudconfigv1.New(c)
	if err != nil {
		panic(err)
	}

	return newCloudConfig
}
