package create

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/giantswarm/k8scloudconfig"
	microerror "github.com/giantswarm/microkit/error"
)

const (
	RoleName                 = "EC2-K8S-Role"
	PolicyName               = "EC2-K8S-Policy"
	ProfileName              = "EC2-DecryptTLSCerts"
	AssumeRolePolicyDocument = `{
		"Version": "2012-10-17",
		"Statement:" {
			"Effect": "Allow",
			"Principal": {
				"Service": "ec2.amazonaws.com",
			},
			"Action": "sts:AssumeRole"
		}
	}`
	PolicyDocumentTempl = `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": "kms:Decrypt",
				"Resource": %q
			}
		]
	}`
)

func (s *Service) encodeTLSAssets(awsSession *session.Session, kmsKeyArn string) (*cloudconfig.CompactTLSAssets, error) {
	rawTLS, err := readRawTLSAssets(s.certsDir)
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	policyDocument := fmt.Sprintf(PolicyDocumentTempl, kmsKeyArn)

	svc := iam.New(awsSession)

	if _, err := svc.CreateRole(&iam.CreateRoleInput{
		RoleName:                 aws.String(RoleName),
		AssumeRolePolicyDocument: aws.String(AssumeRolePolicyDocument),
	}); err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeEntityAlreadyExistsException:
				s.logger.Log("info", fmt.Sprintf("role '%s' already exists, reusing", RoleName))
			default:
				return nil, microerror.MaskAny(err)
			}
		}
	}

	if _, err := svc.PutRolePolicy(&iam.PutRolePolicyInput{
		PolicyName:     aws.String(PolicyName),
		RoleName:       aws.String(RoleName),
		PolicyDocument: aws.String(policyDocument),
	}); err != nil {
		return nil, microerror.MaskAny(err)
	}

	if _, err := svc.CreateInstanceProfile(&iam.CreateInstanceProfileInput{
		InstanceProfileName: aws.String(ProfileName),
	}); err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeEntityAlreadyExistsException:
				s.logger.Log("info", fmt.Sprintf("instance profile '%s' already exists, reusing", RoleName))
			default:
				return nil, microerror.MaskAny(err)
			}
		}
	} else {
		if _, err := svc.AddRoleToInstanceProfile(&iam.AddRoleToInstanceProfileInput{
			InstanceProfileName: aws.String(ProfileName),
			RoleName:            aws.String(RoleName),
		}); err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	encTLS, err := rawTLS.encrypt(awsSession, kmsKeyArn)
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	compTLS, err := encTLS.compact()
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	return compTLS, nil
}
