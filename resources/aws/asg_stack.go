package aws

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	microerror "github.com/giantswarm/microkit/error"
)

type ASGStack struct {
	Client *cloudformation.CloudFormation
	Name   string
	// TemplateURL is the URL of the S3 bucket where the template is stored.
	TemplateURL            string
	SubnetID               string
	AvailabilityZone       string
	ASGMinSize             int
	ASGMaxSize             int
	LoadBalancerName       string
	HealthCheckGracePeriod int
	SecurityGroupID        string
	ImageID                string
	SmallCloudConfig       string
	// IAMInstanceProfileName is the name of the IAM Instance Profile, used to
	// give the instances access to the KMS keys that the CloudConfig is
	// encrypted with.
	IAMInstanceProfileName string
	ClusterID              string
	// KeyName is the name of the EC2 Keypair that contains the SSH key.
	KeyName                  string
	AssociatePublicIPAddress bool
	InstanceType             string
}

func (s *ASGStack) CreateOrFail() error {
	params := &cloudformation.CreateStackInput{
		StackName:   aws.String(s.Name),
		TemplateURL: aws.String(s.TemplateURL),
		Parameters: []*cloudformation.Parameter{
			{
				ParameterKey:   aws.String("SubnetID"),
				ParameterValue: aws.String(s.SubnetID),
			},
			{
				ParameterKey:   aws.String("AZ"),
				ParameterValue: aws.String(s.AvailabilityZone),
			},
			{
				ParameterKey:   aws.String("ASGMinSize"),
				ParameterValue: aws.String(strconv.Itoa(s.ASGMinSize)),
			},
			{
				ParameterKey:   aws.String("ASGMaxSize"),
				ParameterValue: aws.String(strconv.Itoa(s.ASGMaxSize)),
			},
			{
				ParameterKey:   aws.String("LoadBalancerName"),
				ParameterValue: aws.String(s.LoadBalancerName),
			},
			{
				ParameterKey:   aws.String("HealthCheckGracePeriod"),
				ParameterValue: aws.String(strconv.Itoa(s.HealthCheckGracePeriod)),
			},
			{
				ParameterKey:   aws.String("SecurityGroupID"),
				ParameterValue: aws.String(s.SecurityGroupID),
			},
			{
				ParameterKey:   aws.String("SmallCloudConfig"),
				ParameterValue: aws.String(s.SmallCloudConfig),
			},
			{
				ParameterKey:   aws.String("ImageID"),
				ParameterValue: aws.String(s.ImageID),
			},
			{
				ParameterKey:   aws.String("IAMInstanceProfileName"),
				ParameterValue: aws.String(s.IAMInstanceProfileName),
			},
			{
				ParameterKey:   aws.String("KeyName"),
				ParameterValue: aws.String(s.KeyName),
			},
			{
				ParameterKey:   aws.String("AssociatePublicIPAddress"),
				ParameterValue: aws.String(fmt.Sprintf("%t", s.AssociatePublicIPAddress)),
			},
			{
				ParameterKey:   aws.String("InstanceType"),
				ParameterValue: aws.String(s.InstanceType),
			},
		},
		Tags: []*cloudformation.Tag{
			{
				Key:   aws.String(tagKeyName),
				Value: aws.String(s.Name),
			},
			{
				Key:   aws.String(tagKeyCluster),
				Value: aws.String(s.ClusterID),
			},
		},
	}

	if _, err := s.Client.CreateStack(params); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (s *ASGStack) Delete() error {
	if _, err := s.Client.DeleteStack(&cloudformation.DeleteStackInput{
		StackName: aws.String(s.Name),
	}); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}
