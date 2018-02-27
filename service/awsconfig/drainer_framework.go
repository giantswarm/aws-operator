package awsconfig

import (
	"github.com/giantswarm/operatorkit/framework"
)

type DrainerFrameworkConfig struct {
}

func NewDrainerFramework(config DrainerFrameworkConfig) (*framework.Framework, error) {
	return nil, nil
}

func newDrainerResourceRouter(config DrainerFrameworkConfig) (*framework.ResourceRouter, error) {
	return nil, nil
}
