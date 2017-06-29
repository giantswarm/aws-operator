package aws

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/cenkalti/backoff"
	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"
)

// ASGStack represents a CloudFormation stack for an Auto Scaling Group.
type ASGStack struct {
	// Dependencies.
	Client *cloudformation.CloudFormation
	Logger micrologger.Logger

	// Settings.
	ASGMaxSize               int
	ASGMinSize               int
	AssociatePublicIPAddress bool
	AvailabilityZone         string
	ClusterID                string
	HealthCheckGracePeriod   int
	// IAMInstanceProfileName is the name of the IAM Instance Profile, used to
	// give the instances access to the KMS keys that the CloudConfig is
	// encrypted with.
	IAMInstanceProfileName string
	ImageID                string
	InstanceType           string
	// KeyName is the name of the EC2 Keypair that contains the SSH key.
	KeyName          string
	LoadBalancerName string
	Name             string
	SecurityGroupID  string
	SmallCloudConfig string
	SubnetID         string
	// TemplateURL is the URL of the S3 bucket where the CloudFormation template
	// is stored.
	TemplateURL string
}

func (s *ASGStack) CreateOrFail() error {
	params := &cloudformation.CreateStackInput{
		StackName:   aws.String(s.Name),
		TemplateURL: aws.String(s.TemplateURL),
		Parameters: []*cloudformation.Parameter{
			{
				ParameterKey:   aws.String("ASGMinSize"),
				ParameterValue: aws.String(strconv.Itoa(s.ASGMinSize)),
			},
			{
				ParameterKey:   aws.String("ASGMaxSize"),
				ParameterValue: aws.String(strconv.Itoa(s.ASGMaxSize)),
			},
			{
				ParameterKey:   aws.String("AssociatePublicIPAddress"),
				ParameterValue: aws.String(fmt.Sprintf("%t", s.AssociatePublicIPAddress)),
			},
			{
				ParameterKey:   aws.String("AZ"),
				ParameterValue: aws.String(s.AvailabilityZone),
			},
			{
				ParameterKey:   aws.String("HealthCheckGracePeriod"),
				ParameterValue: aws.String(strconv.Itoa(s.HealthCheckGracePeriod)),
			},
			{
				ParameterKey:   aws.String("IAMInstanceProfileName"),
				ParameterValue: aws.String(s.IAMInstanceProfileName),
			},
			{
				ParameterKey:   aws.String("ImageID"),
				ParameterValue: aws.String(s.ImageID),
			},
			{
				ParameterKey:   aws.String("InstanceType"),
				ParameterValue: aws.String(s.InstanceType),
			},
			{
				ParameterKey:   aws.String("KeyName"),
				ParameterValue: aws.String(s.KeyName),
			},
			{
				ParameterKey:   aws.String("LoadBalancerName"),
				ParameterValue: aws.String(s.LoadBalancerName),
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
				ParameterKey:   aws.String("SubnetID"),
				ParameterValue: aws.String(s.SubnetID),
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

	if err := s.waitUntilStackCompleted(s.Name); err != nil {
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

func (s *ASGStack) waitUntilStackCompleted(name string) error {
	listOperation := func() error {
		stacks, err := s.Client.ListStacks(&cloudformation.ListStacksInput{
			StackStatusFilter: []*string{
				aws.String("CREATE_COMPLETE"),
			},
		})
		if err != nil {
			return microerror.MaskAny(err)
		}

		for _, ss := range stacks.StackSummaries {
			if *ss.StackName == s.Name {
				return nil
			}
		}

		return microerror.MaskAnyf(notFoundError, notFoundErrorFormat, "stack", s.Name)
	}

	listNotify := NewNotify(s.Logger, "waiting for stack to be created successfully")
	if err := backoff.RetryNotify(listOperation, NewCustomExponentialBackoff(), listNotify); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}
