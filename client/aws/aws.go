package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type Config struct {
	AccessKeyID     string
	AccessKeySecret string
	Region          string
}

func NewClient(config Config) (*session.Session, *ec2.EC2) {
	awsCfg := &aws.Config{
		Credentials: credentials.NewStaticCredentials(config.AccessKeyID, config.AccessKeySecret, ""),
		Region:      aws.String(config.Region),
	}
	s := session.New(awsCfg)
	ec2client := ec2.New(s)

	return s, ec2client
}
