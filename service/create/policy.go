package create

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

const (
	RoleName                 = "EC2-K8S-Role"
	PolicyName               = "EC2-K8S-Policy"
	ProfileName              = "EC2-K8S-Role"
	AssumeRolePolicyDocument = `{
		"Version": "2012-10-17",
		"Statement": {
			"Effect": "Allow",
			"Principal": {
				"Service": "ec2.amazonaws.com"
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

func createRole(awsSession *session.Session, kmsKeyArn string) error {
	svc := iam.New(awsSession)

	policyDocument := fmt.Sprintf(PolicyDocumentTempl, kmsKeyArn)

	if _, err := svc.CreateRole(&iam.CreateRoleInput{
		RoleName:                 aws.String(RoleName),
		AssumeRolePolicyDocument: aws.String(AssumeRolePolicyDocument),
	}); err != nil {
		return err
	}

	if _, err := svc.PutRolePolicy(&iam.PutRolePolicyInput{
		PolicyName:     aws.String(PolicyName),
		RoleName:       aws.String(RoleName),
		PolicyDocument: aws.String(policyDocument),
	}); err != nil {
		return err
	}

	return nil
}

func createInstanceProfile(awsSession *session.Session) (string, error) {
	svc := iam.New(awsSession)

	if _, err := svc.CreateInstanceProfile(&iam.CreateInstanceProfileInput{
		InstanceProfileName: aws.String(ProfileName),
	}); err != nil {
		return ProfileName, err
	} else {
		if _, err := svc.AddRoleToInstanceProfile(&iam.AddRoleToInstanceProfileInput{
			InstanceProfileName: aws.String(ProfileName),
			RoleName:            aws.String(RoleName),
		}); err != nil {
			return "", err
		}
	}

	if err := svc.WaitUntilInstanceProfileExists(&iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(ProfileName),
	}); err != nil {
		return "", err
	}

	return ProfileName, nil
}
