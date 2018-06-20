package provider

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	awsclient "github.com/giantswarm/e2eclients/aws"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	EC2InstanceType = "m5.xlarge"
)

type AWSConfig struct {
	AWSClient      *awsclient.Client
	GuestFramework *framework.Guest
	HostFramework  *framework.Host
	Logger         micrologger.Logger

	ClusterID string
}

type AWS struct {
	awsClient      *awsclient.Client
	guestFramework *framework.Guest
	hostFramework  *framework.Host
	logger         micrologger.Logger

	clusterID string
}

func NewAWS(config AWSConfig) (*AWS, error) {
	if config.AWSClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.AWSClient must not be empty", config)
	}
	if config.GuestFramework == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.GuestFramework must not be empty", config)
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
		awsClient:      config.AWSClient,
		guestFramework: config.GuestFramework,
		hostFramework:  config.HostFramework,
		logger:         config.Logger,

		clusterID: config.ClusterID,
	}

	return a, nil
}

func (a *AWS) InstallTestApp() error {
	var err error

	var apprClient *apprclient.Client
	{
		c := apprclient.Config{
			Fs:     afero.NewOsFs(),
			Logger: a.logger,

			Address:      CNRAddress,
			Organization: CNROrganization,
		}

		apprClient, err = apprclient.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var helmClient *helmclient.Client
	{
		c := helmclient.Config{
			Logger:    a.logger,
			K8sClient: a.guestFramework.K8sClient(),

			RestConfig: a.guestFramework.RestConfig(),
		}

		helmClient, err = helmclient.New(c)
		if err != nil {
			return microerror.Mask(err)
		}

		err = helmClient.EnsureTillerInstalled()
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// Install the e2e app chart in the guest cluster.
	{
		a.logger.Log("level", "debug", "message", "installing e2e-app for testing")

		tarballPath, err := apprClient.PullChartTarball(ChartName, ChartChannel)
		if err != nil {
			return microerror.Mask(err)
		}

		err = helmClient.InstallFromTarball(tarballPath, ChartNamespace)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
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

func (a *AWS) WaitForAPIDown() error {
	err := a.guestFramework.WaitForAPIDown()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (a *AWS) WaitForGuestReady() error {
	err := a.guestFramework.WaitForGuestReady()
	if err != nil {
		return microerror.Mask(err)
	}

	// Wait for e2e app to be up.
	for {
		a.logger.Log("level", "debug", "message", "waiting for 2 pods of the e2e-app to be up")

		o := metav1.ListOptions{
			LabelSelector: "app=e2e-app",
		}
		l, err := a.guestFramework.K8sClient().CoreV1().Pods(ChartNamespace).List(o)
		if err != nil {
			return microerror.Mask(err)
		}

		if len(l.Items) != 2 {
			a.logger.Log("level", "debug", "message", fmt.Sprintf("found %d pods", len(l.Items)))
			time.Sleep(3 * time.Second)
			continue
		}

		break
	}

	return nil
}
