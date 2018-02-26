package cloudconfigtest

import (
	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-operator/service/awsconfig/v6/cloudconfig"
)

func New() *cloudconfig.CloudConfig {
	c := cloudconfig.DefaultConfig()

	c.Logger = microloggertest.New()

	newCloudConfig, err := cloudconfig.New(c)
	if err != nil {
		panic(err)
	}

	return newCloudConfig
}
