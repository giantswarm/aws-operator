package aws

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type Client struct {
	EC2            *ec2.EC2
	CloudFormation *cloudformation.CloudFormation
}

func NewClient() *Client {
	a := &Client{}

	{
		c := &aws.Config{
			Credentials: credentials.NewStaticCredentials(
				os.Getenv("GUEST_AWS_ACCESS_KEY_ID"),
				os.Getenv("GUEST_AWS_SECRET_ACCESS_KEY"),
				os.Getenv("GUEST_AWS_SESSION_TOKEN")),
			Region: aws.String(os.Getenv("AWS_REGION")),
		}
		a.EC2 = ec2.New(session.New(c))
	}

	{
		c := &aws.Config{
			Credentials: credentials.NewStaticCredentials(
				os.Getenv("HOST_AWS_ACCESS_KEY_ID"),
				os.Getenv("HOST_AWS_SECRET_ACCESS_KEY"),
				os.Getenv("HOST_AWS_SESSION_TOKEN")),
			Region: aws.String(os.Getenv("AWS_REGION")),
		}
		a.CloudFormation = cloudformation.New(session.New(c))
	}

	return a
}
