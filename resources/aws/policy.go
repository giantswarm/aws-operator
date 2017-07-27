package aws

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	microerror "github.com/giantswarm/microkit/error"
)

const (
	RoleNameTemplate         = "EC2-K8S-Role"
	PolicyNameTemplate       = "EC2-K8S-Policy"
	PolicyTemplateFile       = "/opt/templates/%s-iam-policy.tmpl"
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

func (p *Policy) templateName() string {
	return fmt.Sprintf(PolicyTemplateFile, p.PolicyType)
}

func (p *Policy) CreateIfNotExists() (bool, error) {
	return false, fmt.Errorf("instance profiles cannot be reused")
}

func (p *Policy) createRole() error {
	t, err := template.ParseFiles(p.templateName())
	if err != nil {
		return microerror.MaskAny(err)
	}

	var policyDocument bytes.Buffer
	err = t.Execute(&policyDocument, p)
	if err != nil {
		return microerror.MaskAny(err)
	}

	clusterRoleName := p.clusterRoleName()
	if _, err := p.Clients.IAM.CreateRole(&iam.CreateRoleInput{
		RoleName:                 aws.String(clusterRoleName),
		AssumeRolePolicyDocument: aws.String(AssumeRolePolicyDocument),
	}); err != nil {
		return microerror.MaskAny(err)
	}

	if _, err := p.Clients.IAM.PutRolePolicy(&iam.PutRolePolicyInput{
		PolicyName:     aws.String(p.clusterPolicyName()),
		RoleName:       aws.String(clusterRoleName),
		PolicyDocument: aws.String(policyDocument.String()),
	}); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (p *Policy) createInstanceProfile() error {
	if _, err := p.Clients.IAM.CreateInstanceProfile(&iam.CreateInstanceProfileInput{
		InstanceProfileName: aws.String(p.clusterProfileName()),
	}); err != nil {
		return microerror.MaskAny(err)
	} else {
		if _, err := p.Clients.IAM.AddRoleToInstanceProfile(&iam.AddRoleToInstanceProfileInput{
			InstanceProfileName: aws.String(p.clusterProfileName()),
			RoleName:            aws.String(p.clusterRoleName()),
		}); err != nil {
			return microerror.MaskAny(err)
		}
	}

	if err := p.Clients.IAM.WaitUntilInstanceProfileExists(&iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(p.clusterProfileName()),
	}); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (p *Policy) CreateOrFail() error {
	if err := p.createRole(); err != nil {
		return microerror.MaskAny(err)
	}

	if err := p.createInstanceProfile(); err != nil {
		return microerror.MaskAny(err)
	}

	p.name = p.clusterProfileName()

	return nil
}

func (p *Policy) removeRoleFromInstanceProfile() error {
	if _, err := p.Clients.IAM.RemoveRoleFromInstanceProfile(&iam.RemoveRoleFromInstanceProfileInput{
		InstanceProfileName: aws.String(p.clusterProfileName()),
		RoleName:            aws.String(p.clusterRoleName()),
	}); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (p *Policy) deleteInstanceProfile() error {
	if _, err := p.Clients.IAM.DeleteInstanceProfile(&iam.DeleteInstanceProfileInput{
		InstanceProfileName: aws.String(p.clusterProfileName()),
	}); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (p *Policy) deletePolicy() error {
	if _, err := p.Clients.IAM.DeleteRolePolicy(&iam.DeleteRolePolicyInput{
		RoleName:   aws.String(p.clusterRoleName()),
		PolicyName: aws.String(p.clusterPolicyName()),
	}); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (p *Policy) deleteRole() error {
	if _, err := p.Clients.IAM.DeleteRole(&iam.DeleteRoleInput{
		RoleName: aws.String(p.clusterRoleName()),
	}); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (p *Policy) Delete() error {
	if err := p.removeRoleFromInstanceProfile(); err != nil {
		return microerror.MaskAny(err)
	}

	if err := p.deleteInstanceProfile(); err != nil {
		return microerror.MaskAny(err)
	}

	if err := p.deletePolicy(); err != nil {
		return microerror.MaskAny(err)
	}

	if err := p.deleteRole(); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (p Policy) GetName() string {
	return p.name
}
