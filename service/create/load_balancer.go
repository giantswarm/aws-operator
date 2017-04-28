package create

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	awsutil "github.com/giantswarm/aws-operator/client/aws"
	awsresources "github.com/giantswarm/aws-operator/resources/aws"
	"github.com/giantswarm/awstpr"
	microerror "github.com/giantswarm/microkit/error"
)

type LoadBalancerInput struct {
	// Name is the ELB name. It must be unique within a region.
	Name string
	// Clients are the AWS clients.
	Clients awsutil.Clients
	// Cluster is the cluster TPO.
	Cluster awstpr.CustomObject
	// InstanceIDs are the IDs of the instances that should be registered with the ELB.
	InstanceIDs []string
	// PortsToOpen are the ports the ELB should listen to and forward on.
	PortsToOpen awsresources.PortPairs
	// SecurityGroupID is the ID of the security group that will be assigned to the ELB.
	SecurityGroupID string
	// SubnetID is the ID of the subnet the ELB will be placed in.
	SubnetID string
}

func (s *Service) createLoadBalancer(input LoadBalancerInput) (*awsresources.ELB, error) {
	lbName, err := loadBalancerName(input.Name, input.Cluster)
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	lb := &awsresources.ELB{
		Name:          lbName,
		SecurityGroup: input.SecurityGroupID,
		SubnetID:      input.SubnetID,
		PortsToOpen:   input.PortsToOpen,
		Client:        input.Clients.ELB,
	}

	lbCreated, err := lb.CreateIfNotExists()
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	if lbCreated {
		s.logger.Log("debug", fmt.Sprintf("created ELB '%s'", lb.Name))
	} else {
		s.logger.Log("debug", fmt.Sprintf("ELB '%s' already exists, reusing", lb.Name))
	}

	s.logger.Log("debug", "waiting for masters to be ready...")

	var awsFlavouredInstanceIDs []*string
	for _, instanceID := range input.InstanceIDs {
		awsFlavouredInstanceIDs = append(awsFlavouredInstanceIDs, aws.String(instanceID))
	}

	if err := input.Clients.EC2.WaitUntilInstanceRunning(&ec2.DescribeInstancesInput{
		InstanceIds: awsFlavouredInstanceIDs,
	}); err != nil {
		return nil, microerror.MaskAnyf(err, "masters took too long to get running, aborting")
	}

	if err := lb.RegisterInstances(input.InstanceIDs); err != nil {
		return nil, microerror.MaskAnyf(err, "could not register instances with LB: %s")
	}

	s.logger.Log("debug", fmt.Sprintf("instances registered with ELB"))

	return lb, nil
}

func (s *Service) deleteLoadBalancer(input LoadBalancerInput) error {
	// Delete ELB.
	lbName, err := loadBalancerName(input.Name, input.Cluster)
	if err != nil {
		return microerror.MaskAny(err)
	}

	lb := awsresources.ELB{
		Name:   lbName,
		Client: input.Clients.ELB,
	}

	if err := lb.Delete(); err != nil {
		return microerror.MaskAny(err)
	}
	s.logger.Log("debug", fmt.Sprintf("deleted ELB '%s'", lb.Name))

	return nil
}

// loadBalancerName produces a unique name for the load balancer.
// It takes the domain name, extracts the first subdomain, and combines it with the cluster name.
func loadBalancerName(domainName string, cluster awstpr.CustomObject) (string, error) {
	if cluster.Spec.Cluster.Cluster.ID == "" {
		return "", microerror.MaskAnyf(missingCloudConfigKeyError, "spec.cluster.cluster.id")
	}

	componentName, err := componentName(domainName)
	if err != nil {
		return "", microerror.MaskAnyf(malformedCloudConfigKeyError, "spec.cluster.cluster.id")
	}

	lbName := fmt.Sprintf("%s-%s", cluster.Spec.Cluster.Cluster.ID, componentName)

	return lbName, nil
}

// componentName returns the first component of a domain name.
// e.g. apiserver.example.customer.cloud.com -> apiserver
func componentName(domainName string) (string, error) {
	splits := strings.SplitN(domainName, ".", 2)

	if len(splits) != 2 {
		return "", microerror.MaskAny(malformedCloudConfigKeyError)
	}

	return splits[0], nil
}
