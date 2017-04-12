package aws

import (
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/s3"

	microerror "github.com/giantswarm/microkit/error"
)

type Config struct {
	AccessKeyID     string
	AccessKeySecret string
	Region          string
	accountID       string
}

type Clients struct {
	EC2     *ec2.EC2
	IAM     *iam.IAM
	S3      *s3.S3
	KMS     *kms.KMS
	ELB     *elb.ELB
	Route53 *route53.Route53
}

const (
	accountIDPosition = 4
	accountIDLength   = 12
)

func NewClients(config Config) Clients {
	awsCfg := &aws.Config{
		Credentials: credentials.NewStaticCredentials(config.AccessKeyID, config.AccessKeySecret, ""),
		Region:      aws.String(config.Region),
	}
	s := session.New(awsCfg)
	clients := Clients{
		EC2:     ec2.New(s),
		IAM:     iam.New(s),
		S3:      s3.New(s),
		KMS:     kms.New(s),
		ELB:     elb.New(s),
		Route53: route53.New(s),
	}

	return clients
}

func (c *Config) SetAccountID(iamClient *iam.IAM) error {
	resp, err := iamClient.GetUser(&iam.GetUserInput{})
	if err != nil {
		return microerror.MaskAny(err)
	}

	userArn := *resp.User.Arn
	accountID := strings.Split(userArn, ":")[accountIDPosition]

	if err := validateAccountID(accountID); err != nil {
		return microerror.MaskAny(err)
	}

	c.accountID = accountID

	return nil
}

func (c *Config) AccountID() string {
	return c.accountID
}

func validateAccountID(accountID string) error {
	r, _ := regexp.Compile("^[0-9]*$")

	switch {
	case accountID == "":
		return microerror.MaskAny(emptyAmazonAccountIDError)
	case len(accountID) != accountIDLength:
		return microerror.MaskAny(wrongAmazonAccountIDLengthError)
	case !r.MatchString(accountID):
		return microerror.MaskAny(malformedAmazonAccountIDError)
	}

	return nil
}
