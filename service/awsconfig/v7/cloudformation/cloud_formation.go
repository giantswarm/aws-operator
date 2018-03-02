package cloudformation

import (
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
)

type Config struct {
	Client *cloudformation.CloudFormation
}

type CloudFormation struct {
	client *cloudformation.CloudFormation
}

func New(config Config) (*CloudFormation, error) {
	if config.Client == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Client must not be empty", config)
	}

	c := &CloudFormation{
		client: config.Client,
	}

	return c, nil
}

// TODO
func (c *CloudFormation) DescribeOutputs(stackName string) ([]*cloudformation.Output, error) {
	return nil, nil
}

func (c *CloudFormation) GetOutputValue(outputs []*cloudformation.Output, key string) (string, error) {
	for _, o := range outputs {
		if *o.OutputKey == key {
			return *o.OutputValue, nil
		}
	}

	return "", microerror.Maskf(outputNotFoundError, "stack output value for key '%s'", key)
}
