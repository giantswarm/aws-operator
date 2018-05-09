package aws

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v2/resource/cloudformation/adapter"
)

type Config struct {
	AccessKeyID     string
	AccessKeySecret string
	SessionToken    string
	Region          string
	accountID       string
}

type Clients struct {
	AutoScaling    *autoscaling.AutoScaling
	CloudFormation *cloudformation.CloudFormation
	EC2            ec2iface.EC2API
	ELB            elbiface.ELBAPI
	IAM            iamiface.IAMAPI
	KMS            kmsiface.KMSAPI
	Route53        *route53.Route53
	S3             s3iface.S3API
}

const (
	accountIDPosition = 4
	accountIDLength   = 12
)

func NewClients(config Config) Clients {
	s := newSession(config)
	return newClients(s)
}

func NewClientsWithRole(config Config, roleARN string) Clients {
	sess := newSession(config)

	creds := stscreds.NewCredentials(sess, roleARN)

	return newClients(sess, &aws.Config{Credentials: creds})
}

func (c *Config) SetAccountID(iamClient iamiface.IAMAPI) error {
	resp, err := iamClient.GetUser(&iam.GetUserInput{})
	if err != nil {
		return microerror.Mask(err)
	}

	userArn := *resp.User.Arn
	accountID := strings.Split(userArn, ":")[accountIDPosition]

	if err := adapter.ValidateAccountID(accountID); err != nil {
		return microerror.Mask(err)
	}

	c.accountID = accountID

	return nil
}

func (c *Config) AccountID() string {
	return c.accountID
}

func newClients(p client.ConfigProvider, cfgs ...*aws.Config) Clients {
	return Clients{
		AutoScaling:    autoscaling.New(p, cfgs...),
		CloudFormation: cloudformation.New(p, cfgs...),
		EC2:            ec2.New(p, cfgs...),
		ELB:            elb.New(p, cfgs...),
		IAM:            iam.New(p, cfgs...),
		KMS:            kms.New(p, cfgs...),
		Route53:        route53.New(p, cfgs...),
		S3:             s3.New(p, cfgs...),
	}
}

func newSession(config Config) *session.Session {
	awsCfg := &aws.Config{
		Credentials: credentials.NewStaticCredentials(config.AccessKeyID, config.AccessKeySecret, config.SessionToken),
		Region:      aws.String(config.Region),
	}
	return session.New(awsCfg)
}
