package aws

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
)

const (
	// defaultTimeout is the number of minutes after which a Stack creation gets
	// aborted.
	defaultTimeout = 5
)

// ASGStack represents a CloudFormation stack for an Auto Scaling Group.
type ASGStack struct {
	Client *cloudformation.CloudFormation

	// Settings.
	ASGMaxSize               int
	ASGMinSize               int
	ASGType                  string
	AssociatePublicIPAddress bool
	AvailabilityZone         string
	ClusterID                string
	HealthCheckGracePeriod   int
	IAMInstanceProfileName   string
	ImageID                  string
	LoadBalancerName         string
	InstanceType             string
	KeyName                  string
	Name                     string
	SecurityGroupID          string
	SmallCloudConfig         string
	SubnetID                 string
	// TemplateURL is the S3 URL where the CloudFormation template is stored.
	TemplateURL string
	VPCID       string
}

// CreateOrFail creates the autoscaling group stack in Cloud Formation
// or returns the error.
func (s *ASGStack) CreateOrFail() error {
	stackInput := &cloudformation.CreateStackInput{
		StackName:        aws.String(s.Name),
		TemplateURL:      aws.String(s.TemplateURL),
		TimeoutInMinutes: aws.Int64(defaultTimeout),
		Parameters: []*cloudformation.Parameter{
			{
				ParameterKey:   aws.String("ASGMaxSize"),
				ParameterValue: aws.String(strconv.Itoa(s.ASGMaxSize)),
			},
			{
				ParameterKey:   aws.String("ASGMinSize"),
				ParameterValue: aws.String(strconv.Itoa(s.ASGMinSize)),
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
				ParameterKey:   aws.String("LoadBalancerName"),
				ParameterValue: aws.String(s.LoadBalancerName),
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
			{
				ParameterKey:   aws.String("VPCID"),
				ParameterValue: aws.String(s.VPCID),
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

	if _, err := s.Client.CreateStack(stackInput); err != nil {
		return microerror.Mask(err)
	}

	if err := s.Client.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(s.Name),
	}); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// Update updates the autoscaling group stack in Cloud Formation.
func (s *ASGStack) Update() error {
	currentStack, err := s.findExisting()
	if err != nil {
		return microerror.Mask(err)
	}

	updateableParams := map[string]string{
		"ASGMaxSize": strconv.Itoa(s.ASGMaxSize),
		"ASGMinSize": strconv.Itoa(s.ASGMinSize),
		"ImageID":    s.ImageID,
	}
	params := []*cloudformation.Parameter{}

	for _, param := range currentStack.Parameters {
		if value, ok := updateableParams[*param.ParameterKey]; ok {
			param.ParameterValue = aws.String(value)
		}

		params = append(params, param)
	}

	stackInput := &cloudformation.UpdateStackInput{
		Parameters:  params,
		StackName:   aws.String(s.Name),
		TemplateURL: aws.String(s.TemplateURL),
	}
	if _, err := s.Client.UpdateStack(stackInput); err != nil {
		return microerror.Mask(err)
	}

	if err := s.Client.WaitUntilStackUpdateComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(s.Name),
	}); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// Delete deletes the autoscaling group stack in Cloud Formation.
func (s *ASGStack) Delete() error {
	if _, err := s.Client.DeleteStack(&cloudformation.DeleteStackInput{
		StackName: aws.String(s.Name),
	}); err != nil {
		return microerror.Mask(err)
	}

	if err := s.Client.WaitUntilStackDeleteComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(s.Name),
	}); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (s *ASGStack) findExisting() (*cloudformation.Stack, error) {
	stacks, err := s.Client.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(s.Name),
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if len(stacks.Stacks) < 1 {
		return nil, microerror.Maskf(notFoundError, notFoundErrorFormat, CloudFormationStackType, s.Name)
	} else if len(stacks.Stacks) > 1 {
		return nil, microerror.Mask(tooManyResultsError)
	}

	return stacks.Stacks[0], nil
}
