package legacyv2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	awsresources "github.com/giantswarm/aws-operator/resources/aws"
	"github.com/giantswarm/aws-operator/service/keyv2"
)

type LoadBalancerInput struct {
	// Name is the ELB name. It must be unique within a region.
	Name string
	// Clients are the AWS clients.
	Clients awsutil.Clients
	// Cluster is the cluster TPO.
	Cluster v1alpha1.AWSConfig
	// IdleTimeoutSeconds is idle time before closing the front-end and back-end connections
	IdleTimeoutSeconds int
	// InstanceIDs are the IDs of the instances that should be registered with the ELB.
	InstanceIDs []string
	// PortsToOpen are the ports the ELB should listen to and forward on.
	PortsToOpen awsresources.PortPairs
	// SecurityGroupID is the ID of the security group that will be assigned to the ELB.
	SecurityGroupID string
	// SubnetID is the ID of the subnet the ELB will be placed in.
	SubnetID string
	// Scheme, internal for non internet-facing ELBs
	Scheme string
}

func (s *Resource) createLoadBalancer(input LoadBalancerInput) (*awsresources.ELB, error) {
	lbName, err := keyv2.LoadBalancerName(input.Name, input.Cluster)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	lb := &awsresources.ELB{
		Client:             input.Clients.ELB,
		IdleTimeoutSeconds: input.IdleTimeoutSeconds,
		Name:               lbName,
		PortsToOpen:        input.PortsToOpen,
		Scheme:             input.Scheme,
		SecurityGroup:      input.SecurityGroupID,
		SubnetID:           input.SubnetID,
		ClusterID:          keyv2.ClusterID(input.Cluster),
	}

	lbCreated, err := lb.CreateIfNotExists()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if lbCreated {
		s.logger.Log("debug", fmt.Sprintf("created ELB '%s'", lb.Name))

		// Assign the ProxyProtocol policy for ingress controller
		if input.Name == input.Cluster.Spec.Cluster.Kubernetes.IngressController.Domain {
			if err := lb.AssignProxyProtocolPolicy(); err != nil {
				return nil, microerror.Maskf(executionFailedError, fmt.Sprintf("could not assign proxy protocol policy: '%#v'", err))
			}
		}
	} else {
		s.logger.Log("debug", fmt.Sprintf("ELB '%s' already exists, reusing", lb.Name))
	}

	if len(input.InstanceIDs) != 0 {
		if err := s.registerInstances(lb, input); err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return lb, nil
}

func (s *Resource) deleteLoadBalancer(input LoadBalancerInput) error {
	// Delete ELB.
	lbName, err := keyv2.LoadBalancerName(input.Name, input.Cluster)
	if err != nil {
		return microerror.Mask(err)
	}

	lb := awsresources.ELB{
		Name:   lbName,
		Client: input.Clients.ELB,
	}

	if err := lb.Delete(); err != nil {
		return microerror.Mask(err)
	}
	s.logger.Log("debug", fmt.Sprintf("deleted ELB '%s'", lb.Name))

	return nil
}

func (s *Resource) registerInstances(lb *awsresources.ELB, input LoadBalancerInput) error {
	var awsFlavouredInstanceIDs []*string
	for _, instanceID := range input.InstanceIDs {
		awsFlavouredInstanceIDs = append(awsFlavouredInstanceIDs, aws.String(instanceID))
	}

	s.logger.Log("debug", "waiting for instances to be ready...")

	if err := input.Clients.EC2.WaitUntilInstanceRunning(&ec2.DescribeInstancesInput{
		InstanceIds: awsFlavouredInstanceIDs,
	}); err != nil {
		return microerror.Maskf(err, "instances took too long to get running, aborting")
	}

	if err := lb.RegisterInstances(input.InstanceIDs); err != nil {
		return microerror.Maskf(err, "could not register instances with LB: %s")
	}

	s.logger.Log("debug", fmt.Sprintf("instances registered with ELB"))

	return nil
}
