package aws

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
)

const (
	envVarGuestAccessKeyID     = "GUEST_AWS_ACCESS_KEY_ID"
	envVarGuestSecretAccessKey = "GUEST_AWS_SECRET_ACCESS_KEY"
	envVarGuestSessionToken    = "GUEST_AWS_SESSION_TOKEN"
	envVarHostAccessKeyID      = "HOST_AWS_ACCESS_KEY_ID"
	envVarHostSecretAccessKey  = "HOST_AWS_SECRET_ACCESS_KEY"
	envVarHostSessionToken     = "HOST_AWS_SESSION_TOKEN"
	envVarRegion               = "AWS_REGION"
)

var (
	guestAccessKeyID     string
	guestSecretAccessKey string
	guestSessionToken    string
	hostAccessKeyID      string
	hostSecretAccessKey  string
	hostSessionToken     string
	region               string
)

type Client struct {
	EC2            *ec2.EC2
	CloudFormation *cloudformation.CloudFormation
}

func NewClient() (*Client, error) {
	a := &Client{}

	{
		region = os.Getenv(envVarRegion)
		if region == "" {
			return nil, microerror.Maskf(invalidConfigError, "%s must be set", envVarRegion)
		}
	}

	{
		guestAccessKeyID = os.Getenv(envVarGuestAccessKeyID)
		if guestAccessKeyID == "" {
			return nil, microerror.Maskf(invalidConfigError, "%s must be set", envVarGuestAccessKeyID)
		}

		guestSecretAccessKey = os.Getenv(envVarGuestSecretAccessKey)
		if guestSecretAccessKey == "" {
			return nil, microerror.Maskf(invalidConfigError, "%s must be set", envVarGuestSecretAccessKey)
		}

		guestSessionToken = os.Getenv(envVarGuestSessionToken)
		c := &aws.Config{
			Credentials: credentials.NewStaticCredentials(
				guestAccessKeyID,
				guestSecretAccessKey,
				guestSessionToken),
			Region: aws.String(region),
		}
		a.EC2 = ec2.New(session.New(c))
	}

	{
		hostAccessKeyID = os.Getenv(envVarHostAccessKeyID)
		if hostAccessKeyID == "" {
			return nil, microerror.Maskf(invalidConfigError, "%s must be set", envVarHostAccessKeyID)
		}

		hostSecretAccessKey = os.Getenv(envVarHostSecretAccessKey)
		if hostSecretAccessKey == "" {
			return nil, microerror.Maskf(invalidConfigError, "%s must be set", envVarHostSecretAccessKey)
		}

		hostSessionToken = os.Getenv(envVarHostSessionToken)
		c := &aws.Config{
			Credentials: credentials.NewStaticCredentials(
				hostAccessKeyID,
				hostSecretAccessKey,
				hostSessionToken),
			Region: aws.String(region),
		}
		a.CloudFormation = cloudformation.New(session.New(c))
	}

	return a, nil
}
