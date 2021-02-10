package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/elbv2/elbv2iface"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/servicequotas"
	"github.com/aws/aws-sdk-go/service/servicequotas/servicequotasiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/aws/aws-sdk-go/service/support"
	"github.com/aws/aws-sdk-go/service/support/supportiface"
	"github.com/giantswarm/microerror"
)

const (
	// trustedAdvisorRegion describes the AWS region in which the trusted advisor
	// service is available.
	trustedAdvisorRegion = "us-east-1"
)

type Config struct {
	AccessKeyID     string
	AccessKeySecret string
	Region          string
	RoleARN         string
	SessionToken    string
}

type Clients struct {
	AutoScaling    *autoscaling.AutoScaling
	CloudFormation *cloudformation.CloudFormation
	EC2            ec2iface.EC2API
	ELB            elbiface.ELBAPI
	ELBv2          elbv2iface.ELBV2API
	IAM            iamiface.IAMAPI
	KMS            kmsiface.KMSAPI
	Route53        *route53.Route53
	S3             s3iface.S3API
	ServiceQuotas  servicequotasiface.ServiceQuotasAPI
	STS            stsiface.STSAPI
	Support        supportiface.SupportAPI
}

func NewClients(config Config) (Clients, error) {
	if config.AccessKeyID == "" {
		return Clients{}, microerror.Maskf(invalidConfigError, "%T.AccessKeyID must not be empty", config)
	}
	if config.AccessKeySecret == "" {
		return Clients{}, microerror.Maskf(invalidConfigError, "%T.AccessKeySecret must not be empty", config)
	}
	if config.Region == "" {
		return Clients{}, microerror.Maskf(invalidConfigError, "%T.Region must not be empty", config)
	}
	if config.RoleARN == "" {
		return Clients{}, microerror.Maskf(invalidConfigError, "%T.RoleARN must not be empty", config)
	}

	var err error

	var s *session.Session
	{
		c := &aws.Config{
			Credentials: credentials.NewStaticCredentials(config.AccessKeyID, config.AccessKeySecret, config.SessionToken),
			Region:      aws.String(config.Region),
		}

		s, err = session.NewSession(c)
		if err != nil {
			return Clients{}, microerror.Mask(err)
		}
	}

	c := newClients(s, config.RoleARN)
	return c, nil
}

func newClients(session *session.Session, roleARN string) Clients {
	credentialsConfig := &aws.Config{
		Credentials: stscreds.NewCredentials(session, roleARN),
	}
	supportConfig := aws.NewConfig().WithRegion(trustedAdvisorRegion)

	c := Clients{
		AutoScaling:    autoscaling.New(session, credentialsConfig),
		CloudFormation: cloudformation.New(session, credentialsConfig),
		EC2:            ec2.New(session, credentialsConfig),
		ELB:            elb.New(session, credentialsConfig),
		ELBv2:          elbv2.New(session, credentialsConfig),
		IAM:            iam.New(session, credentialsConfig),
		KMS:            kms.New(session, credentialsConfig),
		Route53:        route53.New(session, credentialsConfig),
		S3:             s3.New(session, credentialsConfig),
		ServiceQuotas:  servicequotas.New(session, credentialsConfig),
		STS:            sts.New(session, credentialsConfig),
		Support:        support.New(session, credentialsConfig, supportConfig),
	}

	return c
}
