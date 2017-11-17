package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/microerror"
)

const (
	RoleNameTemplate         = "EC2-K8S-Role"
	PolicyNameTemplate       = "EC2-K8S-Policy"
	ProfileNameTemplate      = "EC2-K8S-Role"
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
	MasterPolicyDocumentTempl = `{
		"Version": "2012-10-17",
		"Statement": [
			{
            	"Action": "ec2:*",
            	"Effect": "Allow",
                "Resource": "*"
            },
			{
				"Effect": "Allow",
				"Action": "kms:Decrypt",
				"Resource": %q
			},
			{
				"Effect": "Allow",
				"Action": [
					"s3:GetBucketLocation",
					"s3:ListAllMyBuckets"
				],
				"Resource": "*"
			},
			{
				"Effect": "Allow",
				"Action": [
					"s3:ListBucket"
				],
				"Resource": "arn:aws:s3:::%s"
			},
			{
				"Effect": "Allow",
				"Action": "s3:GetObject",
				"Resource": "arn:aws:s3:::%s/*"
			},
			{
				"Effect": "Allow",
				"Action": "elasticloadbalancing:*",
				"Resource": "*"
			}
		]
	}`
	MasterPolicyType          = "master"
	WorkerPolicyDocumentTempl = `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": "ec2:Describe*",
				"Resource": "*"
			},
			{
				"Effect": "Allow",
				"Action": "ec2:AttachVolume",
				"Resource": "*"
			},
			{
				"Effect": "Allow",
				"Action": "ec2:DetachVolume",
				"Resource": "*"
			},
			{
				"Effect": "Allow",
				"Action": "kms:Decrypt",
				"Resource": %q
			},
			{
				"Effect": "Allow",
				"Action": [
					"s3:GetBucketLocation",
					"s3:ListAllMyBuckets"
				],
				"Resource": "*"
			},
			{
				"Effect": "Allow",
				"Action": [
					"s3:ListBucket"
				],
				"Resource": "arn:aws:s3:::%s"
			},
			{
				"Effect": "Allow",
				"Action": "s3:GetObject",
				"Resource": "arn:aws:s3:::%s/*"
			},
			{
				"Effect": "Allow",
				"Action": [
					"ecr:GetAuthorizationToken",
					"ecr:BatchCheckLayerAvailability",
					"ecr:GetDownloadUrlForLayer",
					"ecr:GetRepositoryPolicy",
					"ecr:DescribeRepositories",
					"ecr:ListImages",
					"ecr:BatchGetImage"
				],
				"Resource": "*"
			}
		]
	}`
	WorkerPolicyType = "worker"
)

type Policy struct {
	ClusterID  string
	KMSKeyArn  string
	PolicyType string
	S3Bucket   string
	name       string
	AWSEntity
}

func (p *Policy) clusterPolicyName() string {
	return fmt.Sprintf("%s-%s-%s", p.ClusterID, p.PolicyType, PolicyNameTemplate)
}

func (p *Policy) clusterProfileName() string {
	return fmt.Sprintf("%s-%s-%s", p.ClusterID, p.PolicyType, ProfileNameTemplate)
}

func (p *Policy) clusterRoleName() string {
	return fmt.Sprintf("%s-%s-%s", p.ClusterID, p.PolicyType, RoleNameTemplate)
}

func (p *Policy) CreateIfNotExists() (bool, error) {
	err := p.CreateOrFail()
	if err == nil {
		return true, nil
	}
	if awsutil.IsIAMRoleDuplicateError(err) {
		return false, nil
	}
	return false, microerror.Mask(err)
}

func (p *Policy) createRole() error {
	// TODO switch to using a file and Go templates
	var policyTemplate string

	switch p.PolicyType {
	case MasterPolicyType:
		policyTemplate = MasterPolicyDocumentTempl
	case WorkerPolicyType:
		policyTemplate = WorkerPolicyDocumentTempl
	default:
		return microerror.Maskf(notFoundError, notFoundErrorFormat, "PolicyType", p.PolicyType)
	}

	policyDocument := fmt.Sprintf(policyTemplate, p.KMSKeyArn, p.S3Bucket, p.S3Bucket)
	clusterRoleName := p.clusterRoleName()
	if _, err := p.Clients.IAM.CreateRole(&iam.CreateRoleInput{
		RoleName:                 aws.String(clusterRoleName),
		AssumeRolePolicyDocument: aws.String(AssumeRolePolicyDocument),
	}); err != nil {
		return microerror.Mask(err)
	}

	if _, err := p.Clients.IAM.PutRolePolicy(&iam.PutRolePolicyInput{
		PolicyName:     aws.String(p.clusterPolicyName()),
		RoleName:       aws.String(clusterRoleName),
		PolicyDocument: aws.String(policyDocument),
	}); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (p *Policy) createInstanceProfile() error {
	if _, err := p.Clients.IAM.CreateInstanceProfile(&iam.CreateInstanceProfileInput{
		InstanceProfileName: aws.String(p.clusterProfileName()),
	}); err != nil {
		return microerror.Mask(err)
	} else {
		if _, err := p.Clients.IAM.AddRoleToInstanceProfile(&iam.AddRoleToInstanceProfileInput{
			InstanceProfileName: aws.String(p.clusterProfileName()),
			RoleName:            aws.String(p.clusterRoleName()),
		}); err != nil {
			return microerror.Mask(err)
		}
	}

	if err := p.Clients.IAM.WaitUntilInstanceProfileExists(&iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(p.clusterProfileName()),
	}); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (p *Policy) CreateOrFail() error {
	if err := p.createRole(); err != nil {
		return microerror.Mask(err)
	}

	if err := p.createInstanceProfile(); err != nil {
		return microerror.Mask(err)
	}

	p.name = p.clusterProfileName()

	return nil
}

func (p *Policy) removeRoleFromInstanceProfile() error {
	if _, err := p.Clients.IAM.RemoveRoleFromInstanceProfile(&iam.RemoveRoleFromInstanceProfileInput{
		InstanceProfileName: aws.String(p.clusterProfileName()),
		RoleName:            aws.String(p.clusterRoleName()),
	}); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (p *Policy) deleteInstanceProfile() error {
	if _, err := p.Clients.IAM.DeleteInstanceProfile(&iam.DeleteInstanceProfileInput{
		InstanceProfileName: aws.String(p.clusterProfileName()),
	}); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (p *Policy) deletePolicy() error {
	if _, err := p.Clients.IAM.DeleteRolePolicy(&iam.DeleteRolePolicyInput{
		RoleName:   aws.String(p.clusterRoleName()),
		PolicyName: aws.String(p.clusterPolicyName()),
	}); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (p *Policy) deleteRole() error {
	if _, err := p.Clients.IAM.DeleteRole(&iam.DeleteRoleInput{
		RoleName: aws.String(p.clusterRoleName()),
	}); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (p *Policy) Delete() error {
	if err := p.removeRoleFromInstanceProfile(); err != nil {
		return microerror.Mask(err)
	}

	if err := p.deleteInstanceProfile(); err != nil {
		return microerror.Mask(err)
	}

	if err := p.deletePolicy(); err != nil {
		return microerror.Mask(err)
	}

	if err := p.deleteRole(); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (p Policy) GetName() string {
	return p.name
}
