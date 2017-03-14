package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Config struct {
	AccessKeyID     string
	AccessKeySecret string
	Region          string
}

type Clients struct {
	EC2 *ec2.EC2
	IAM *iam.IAM
	S3  *s3.S3
	KMS *kms.KMS
}

func NewClients(config Config) Clients {
	awsCfg := &aws.Config{
		Credentials: credentials.NewStaticCredentials(config.AccessKeyID, config.AccessKeySecret, ""),
		Region:      aws.String(config.Region),
	}
	s := session.New(awsCfg)
	clients := Clients{
		EC2: ec2.New(s),
		IAM: iam.New(s),
		S3:  s3.New(s),
		KMS: kms.New(s),
	}

	return clients
}
