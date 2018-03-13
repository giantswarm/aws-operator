// +build k8srequired

package client

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// TODO it would be best if this would be aligned with the client we use in the
// aws-operator. The feel would be more native and common and less differences
// makes it easier to understand certain internals.
type AWS struct {
	EC2            *ec2.EC2
	CloudFormation *cloudformation.CloudFormation
}

func NewAWS() *AWS {
	a := &AWS{}

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
