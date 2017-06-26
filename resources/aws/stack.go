package aws

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	microerror "github.com/giantswarm/microkit/error"
)

type Stack struct {
	Client *cloudformation.CloudFormation
	Name   string
	// TemplateURL is the URL of the S3 bucket where the template is stored.
	TemplateURL             string
	SubnetID                string
	AvailabilityZone        string
	ASGMinSize              int
	ASGMaxSize              int
	LaunchConfigurationName string
	LoadBalancerName        string
	HealthCheckGracePeriod  int
	SecurityGroupID         string
	ImageID                 string
	SmallCloudConfig        string
}

func (s *Stack) CreateOrFail() error {
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
				ParameterKey:   aws.String("WorkersLaunchConfigurationName"),
				ParameterValue: aws.String(s.LaunchConfigurationName),
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
				ParameterKey:   aws.String("SecurityGroupName"),
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
		},
	}

	fmt.Printf("params: %+v\n", params)

	if _, err := s.Client.CreateStack(params); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (s *Stack) Delete() error {
	if _, err := s.Client.DeleteStack(&cloudformation.DeleteStackInput{
		StackName: aws.String(s.Name),
	}); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}
