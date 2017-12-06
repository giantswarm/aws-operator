package adapter

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/aws-operator/service/keyv1"
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/microerror"
)

// template related to this adapter: service/templates/cloudformation/auto_scaling_group.go

type autoScalingGroupAdapter struct {
	ASGMaxSize             int
	ASGMinSize             int
	AZ                     string
	HealthCheckGracePeriod int
	LoadBalancerName       string
	MaxBatchSize           string
	MinInstancesInService  string
	RollingUpdatePauseTime string
	SubnetID               string
	ClusterID              string
}

func (a *autoScalingGroupAdapter) getAutoScalingGroup(customObject awstpr.CustomObject, clients Clients) error {
	a.AZ = customObject.Spec.AWS.AZ
	workers := keyv1.WorkerCount(customObject)
	a.ASGMaxSize = workers
	a.ASGMinSize = workers
	a.MaxBatchSize = strconv.FormatFloat(asgMaxBatchSizeRatio, 'f', -1, 32)
	a.MinInstancesInService = strconv.FormatFloat(asgMinInstancesRatio, 'f', -1, 32)
	a.HealthCheckGracePeriod = gracePeriodSeconds
	a.RollingUpdatePauseTime = rollingUpdatePauseTime

	// load balancer name
	// TODO: remove this code once the ingress load balancer is created by cloudformation
	// and add a reference in the template
	lbName, err := keyv1.LoadBalancerName(customObject.Spec.Cluster.Kubernetes.IngressController.Domain, customObject)
	if err != nil {
		return microerror.Mask(err)
	}
	a.LoadBalancerName = lbName

	// subnet ID
	// TODO: remove this code once the subnet is created by cloudformation and add a
	// reference in the template
	subnetName := keyv1.SubnetName(customObject, suffixPublic)
	describeSubnetInput := &ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String(fmt.Sprintf("tag:%s", tagKeyName)),
				Values: []*string{
					aws.String(subnetName),
				},
			},
		},
	}
	output, err := clients.EC2.DescribeSubnets(describeSubnetInput)
	if err != nil {
		return microerror.Mask(err)
	}
	if len(output.Subnets) > 1 {
		return microerror.Mask(tooManyResultsError)
	}

	a.SubnetID = *output.Subnets[0].SubnetId

	a.ClusterID = keyv1.ClusterID(customObject)

	return nil
}
