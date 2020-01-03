package aws

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
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
	CloudFormation *cloudformation.CloudFormation
	EC2            *ec2.EC2
	S3             *s3.S3
}

func NewClient() (*Client, error) {
	a := &Client{}

	{
		region = os.Getenv(envVarRegion)
		if region == "" {
			return nil, microerror.Maskf(invalidConfigError, "%s must be set", envVarRegion)
		}
	}

	var hostSession *session.Session
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
		config := &aws.Config{
			Credentials: credentials.NewStaticCredentials(
				hostAccessKeyID,
				hostSecretAccessKey,
				hostSessionToken),
			Region: aws.String(region),
		}
		var err error
		hostSession, err = session.NewSession(config)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var guestSession *session.Session
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
		config := &aws.Config{
			Credentials: credentials.NewStaticCredentials(
				guestAccessKeyID,
				guestSecretAccessKey,
				guestSessionToken),
			Region: aws.String(region),
		}
		var err error
		guestSession, err = session.NewSession(config)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	a.CloudFormation = cloudformation.New(hostSession)
	a.EC2 = ec2.New(guestSession)
	a.S3 = s3.New(guestSession)

	return a, nil
}
