package aws

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
)

const (
	// asgMaxSizeParam is the Cloud Formation parameter name for the max ASG size.
	asgMaxSizeParam = "ASGMaxSize"
	// asgMinSizeParam is the Cloud Formation parameter name for the min ASG size.
	asgMinSizeParam = "ASGMinSize"
	// defaultTimeout is the number of minutes after which a Stack creation gets
	// aborted.
	defaultTimeout = 5
	// imageIDParam is the Cloud Formation parameter name for the ASG image ID.
	imageIDParam = "ImageID"

	stackDoesNotExistError = "Stack with id %s does not exist"
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
				ParameterKey:   aws.String(asgMaxSizeParam),
				ParameterValue: aws.String(strconv.Itoa(s.ASGMaxSize)),
			},
			{
				ParameterKey:   aws.String(asgMinSizeParam),
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
				ParameterKey:   aws.String(imageIDParam),
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

// Update updates the autoscaling group stack in Cloud Formation if one of the
// updatable parameters has changed.
func (s *ASGStack) Update() error {
	currentStack, err := s.findExisting()
	if err != nil {
		return microerror.Mask(err)
	}

	updateableParams := map[string]string{
		asgMaxSizeParam: strconv.Itoa(s.ASGMaxSize),
		asgMinSizeParam: strconv.Itoa(s.ASGMinSize),
		imageIDParam:    s.ImageID,
	}

	if hasStackChanged(currentStack.Parameters, updateableParams) {
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

// CheckIfExists checks if there is an autoscaling group stack in Cloud Formation
// with the provided name.
func (s *ASGStack) CheckIfExists() (bool, error) {
	_, err := s.findExisting()
	if IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func (s *ASGStack) findExisting() (*cloudformation.Stack, error) {
	stacks, err := s.Client.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(s.Name),
	})
	if err != nil {
		underlying := microerror.Cause(err)
		if awserr, ok := underlying.(awserr.Error); ok {
			if awserr.Message() == fmt.Sprintf(stackDoesNotExistError, s.Name) {
				return nil, microerror.Mask(notFoundError)
			}
		}

		return nil, microerror.Mask(err)
	}

	if len(stacks.Stacks) < 1 {
		return nil, microerror.Maskf(notFoundError, notFoundErrorFormat, CloudFormationStackType, s.Name)
	} else if len(stacks.Stacks) > 1 {
		return nil, microerror.Mask(tooManyResultsError)
	}

	return stacks.Stacks[0], nil
}

func hasStackChanged(params []*cloudformation.Parameter, paramUpdates map[string]string) bool {
	for _, param := range params {
		key := *param.ParameterKey
		currentValue := *param.ParameterValue

		if updatedValue, ok := paramUpdates[key]; ok {
			if updatedValue != currentValue {
				return true
			}
		}
	}

	return false
}
