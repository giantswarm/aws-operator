package provider

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	awsclient "github.com/giantswarm/e2eclients/aws"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	EC2InstanceType = "m4.xlarge"
)

type AWSConfig struct {
	AWSClient     *awsclient.Client
	HostFramework *framework.Host
	Logger        micrologger.Logger

	ClusterID string
}

type AWS struct {
	awsClient     *awsclient.Client
	hostFramework *framework.Host
	logger        micrologger.Logger

	clusterID string
}

func NewAWS(config AWSConfig) (*AWS, error) {
	if config.AWSClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.AWSClient must not be empty", config)
	}
	if config.HostFramework == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HostFramework must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.ClusterID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterID must not be empty", config)
	}

	a := &AWS{
		awsClient:     config.AWSClient,
		hostFramework: config.HostFramework,
		logger:        config.Logger,

		clusterID: config.ClusterID,
	}

	return a, nil
}

func (a *AWS) RebootMaster() error {
	masterInstanceName := fmt.Sprintf("%s-master", a.clusterID)
	describeInput := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:Name"),
				Values: []*string{
					aws.String(masterInstanceName),
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
	res, err := a.awsClient.EC2.DescribeInstances(describeInput)
	if err != nil {
		return microerror.Mask(err)
	}
	if len(res.Reservations) == 0 {
		return microerror.Maskf(notFoundError, "ec2 instance %q not found", masterInstanceName)
	}
	if len(res.Reservations) > 1 {
		return microerror.Maskf(tooManyResultsError, "found %d running instances named %q", len(res.Reservations), masterInstanceName)
	}

	masterID := res.Reservations[0].Instances[0].InstanceId
	rebootInput := &ec2.RebootInstancesInput{
		InstanceIds: []*string{
			masterID,
		},
	}
	_, err = a.awsClient.EC2.RebootInstances(rebootInput)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (a *AWS) ReplaceMaster() error {
	customObject, err := a.hostFramework.G8sClient().ProviderV1alpha1().AWSConfigs("default").Get(a.clusterID, metav1.GetOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	// Change instance type to trigger replacement of existing master node.
	customObject.Spec.AWS.Masters[0].InstanceType = EC2InstanceType

	_, err = a.hostFramework.G8sClient().ProviderV1alpha1().AWSConfigs("default").Update(customObject)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
