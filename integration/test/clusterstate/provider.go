package clusterstate

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	awsclient "github.com/giantswarm/e2eclients/aws"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	EC2InstanceType = "m4.xlarge"
)

type ProviderConfig struct {
	AWSClient *awsclient.Client
	G8sClient versioned.Interface
	Logger    micrologger.Logger

	ClusterID string
}

type Provider struct {
	awsClient *awsclient.Client
	g8sClient versioned.Interface
	logger    micrologger.Logger

	clusterID string
}

func NewProvider(config ProviderConfig) (*Provider, error) {
	if config.AWSClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.AWSClient must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.ClusterID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterID must not be empty", config)
	}

	p := &Provider{
		awsClient: config.AWSClient,
		g8sClient: config.G8sClient,
		logger:    config.Logger,

		clusterID: config.ClusterID,
	}

	return p, nil
}

func (p *Provider) RebootMaster() error {
	var instanceID string
	{
		i := &ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("tag:Name"),
					Values: []*string{
						aws.String(fmt.Sprintf("%s-master", p.clusterID)),
					},
				},
				{
					Name: aws.String("instance-state-name"),
					Values: []*string{
						aws.String("running"),
					},
				},
			},
		}

		o, err := p.awsClient.EC2.DescribeInstances(i)
		if err != nil {
			return microerror.Mask(err)
		}
		if len(o.Reservations) != 1 {
			return microerror.Maskf(executionFailedError, "expected one master instance, got %d", len(o.Reservations))
		}
		if len(o.Reservations[0].Instances) != 1 {
			return microerror.Maskf(executionFailedError, "expected one master instance, got %d", len(o.Reservations[0].Instances))
		}

		instanceID = *o.Reservations[0].Instances[0].InstanceId
	}

	{
		i := &ec2.RebootInstancesInput{
			InstanceIds: []*string{
				aws.String(instanceID),
			},
		}

		_, err := p.awsClient.EC2.RebootInstances(i)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (p *Provider) ReplaceMaster() error {
	customObject, err := p.g8sClient.ProviderV1alpha1().AWSConfigs("default").Get(p.clusterID, metav1.GetOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	// Change instance type to trigger replacement of existing master node.
	customObject.Spec.AWS.Masters[0].InstanceType = EC2InstanceType

	_, err = p.g8sClient.ProviderV1alpha1().AWSConfigs("default").Update(customObject)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
