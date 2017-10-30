package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/giantswarm/microerror"
)

// ELB is an Elastic Load Balancer
type ELB struct {
	AZ                 string
	Client             *elb.ELB
	dnsName            string
	hostedZoneID       string
	IdleTimeoutSeconds int
	Name               string
	PortsToOpen        PortPairs
	Scheme             string
	SecurityGroup      string
	SubnetID           string
	Tags               []string
}

// PortPair is a pair of ports.
type PortPair struct {
	// PortELB is the port the ELB should listen on.
	PortELB int
	// PortInstance is the port on the instance the ELB forwards traffic to.
	PortInstance int
}

// PortPairs is an array of PortPair.
type PortPairs []PortPair

const (
	// proxyProtocolPolicyTypeName is the name of the ProxyProtocolPolicy type.
	proxyProtocolPolicyTypeName = "ProxyProtocolPolicyType"
	// proxyProtocolPolicyNameSuffix is the suffix we use for the name of our ProxyProtocol policy.
	proxyProtocolPolicyNameSuffix = "proxy-protocol-policy"
	// proxyProtocolAttributeName is the name of the ProxyProtocol attribute we set on the policy.
	proxyProtocolAttributeName = "ProxyProtocol"
	// Default values for health checks.
	healthCheckHealthyThreshold   = 10
	healthCheckInterval           = 5
	healthCheckTimeout            = 3
	healthCheckUnhealthyThreshold = 2
	// Default value for connections, AWS default is 60
	idleTimeoutSeconds = 60
)

func (lb *ELB) CreateIfNotExists() (bool, error) {
	if lb.Client == nil {
		return false, microerror.Mask(clientNotInitializedError)
	}

	_, err := lb.findExisting()
	if err == nil {
		return false, nil
	}

	if !strings.Contains(err.Error(), notFoundError.Error()) {

		return false, err
	}

	if err := lb.CreateOrFail(); err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func (lb *ELB) CreateOrFail() error {
	if lb.Client == nil {
		return microerror.Mask(clientNotInitializedError)
	}
	if len(lb.PortsToOpen) == 0 {
		return microerror.Maskf(attributeEmptyError, attributeEmptyErrorFormat, "portsToOpen")
	}
	if lb.IdleTimeoutSeconds <= 0 {
		// Set to default if zero (struct initialization) or negative (invalid config)
		lb.IdleTimeoutSeconds = idleTimeoutSeconds
	}

	var listeners []*elb.Listener
	for _, portPair := range lb.PortsToOpen {
		listener := &elb.Listener{
			InstancePort:     aws.Int64(int64(portPair.PortInstance)),
			LoadBalancerPort: aws.Int64(int64(portPair.PortELB)),
			// We use TCP and not HTTP(S) because we want to do SSL passthrough and not termination.
			Protocol: aws.String("TCP"),
		}

		listeners = append(listeners, listener)
	}

	if _, err := lb.Client.CreateLoadBalancer(&elb.CreateLoadBalancerInput{
		LoadBalancerName: aws.String(lb.Name),
		Listeners:        listeners,
		SecurityGroups: []*string{
			aws.String(lb.SecurityGroup),
		},
		Subnets: []*string{
			aws.String(lb.SubnetID),
		},
		Scheme: aws.String(lb.Scheme),
	}); err != nil {
		return microerror.Mask(err)
	}

	if _, err := lb.Client.ConfigureHealthCheck(&elb.ConfigureHealthCheckInput{
		HealthCheck: &elb.HealthCheck{
			HealthyThreshold:   aws.Int64(int64(healthCheckHealthyThreshold)),
			Interval:           aws.Int64(int64(healthCheckInterval)),
			Target:             aws.String(fmt.Sprintf("TCP:%d", lb.PortsToOpen[0].PortInstance)),
			Timeout:            aws.Int64(int64(healthCheckTimeout)),
			UnhealthyThreshold: aws.Int64(int64(healthCheckUnhealthyThreshold)),
		},
		LoadBalancerName: aws.String(lb.Name),
	}); err != nil {
		return microerror.Mask(err)
	}

	// Configure additional attributes
	if _, err := lb.Client.ModifyLoadBalancerAttributes(&elb.ModifyLoadBalancerAttributesInput{
		LoadBalancerAttributes: &elb.LoadBalancerAttributes{
			ConnectionSettings: &elb.ConnectionSettings{
				IdleTimeout: aws.Int64(int64(lb.IdleTimeoutSeconds)),
			},
		},
		LoadBalancerName: aws.String(lb.Name),
	}); err != nil {
		return microerror.Mask(err)
	}

	// We have to populate some additional fields.
	lbDescription, err := lb.findExisting()
	if err != nil {
		return microerror.Mask(err)
	}

	lb.setDNSFields(*lbDescription)

	return nil
}

func (lb ELB) Delete() error {
	if lb.Client == nil {
		return microerror.Mask(clientNotInitializedError)
	}

	if _, err := lb.Client.DeleteLoadBalancer(&elb.DeleteLoadBalancerInput{
		LoadBalancerName: aws.String(lb.Name),
	}); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (lb *ELB) RegisterInstances(instanceIDs []string) error {
	var instances []*elb.Instance

	for _, id := range instanceIDs {
		elbInstance := &elb.Instance{
			InstanceId: aws.String(id),
		}
		instances = append(instances, elbInstance)
	}

	if _, err := lb.Client.RegisterInstancesWithLoadBalancer(&elb.RegisterInstancesWithLoadBalancerInput{
		Instances:        instances,
		LoadBalancerName: aws.String(lb.Name),
	}); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// AssignProxyProtocolPolicy creates a ProxyProtocol policy and assigns it to the Load Balancer.
// This is needed for ELBs that listen/forward over TCP, in order to add
// a header with the address, port of the source and destination.
// Without this, `kubectl log/exec` don't work.
// See https://github.com/kubernetes/ingress/tree/4601775c18f5c6968e56e1eeaa26efc629590bb0/controllers/nginx#proxy-protocol
func (lb *ELB) AssignProxyProtocolPolicy() error {
	policyName := fmt.Sprintf("%s-%s", lb.Name, proxyProtocolPolicyNameSuffix)

	if _, err := lb.Client.CreateLoadBalancerPolicy(&elb.CreateLoadBalancerPolicyInput{
		LoadBalancerName: aws.String(lb.Name),
		PolicyName:       aws.String(policyName),
		PolicyTypeName:   aws.String(proxyProtocolPolicyTypeName),
		PolicyAttributes: []*elb.PolicyAttribute{
			{
				AttributeName:  aws.String(proxyProtocolAttributeName),
				AttributeValue: aws.String("true"),
			},
		},
	}); err != nil {
		return microerror.Mask(err)
	}

	setPolicyInput := &elb.SetLoadBalancerPoliciesForBackendServerInput{
		LoadBalancerName: aws.String(lb.Name),
		PolicyNames:      []*string{aws.String(policyName)},
	}
	for _, portPair := range lb.PortsToOpen {
		setPolicyInput.InstancePort = aws.Int64(int64(portPair.PortInstance))

		if _, err := lb.Client.SetLoadBalancerPoliciesForBackendServer(setPolicyInput); err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (lb ELB) DNSName() string {
	return lb.dnsName
}

func (lb ELB) HostedZoneID() string {
	return lb.hostedZoneID
}

// NewELBFromExisting initializes an ELB struct with some fields retrieved from the API,
// such as its FQDN and its Hosted Zone ID. We need these fields when deleting a Record Set.
// This method doesn't create a new ELB on AWS.
func NewELBFromExisting(name string, client *elb.ELB) (*ELB, error) {
	lb := ELB{
		Name:   name,
		Client: client,
	}

	lbDescription, err := lb.findExisting()
	if err != nil {
		return nil, err
	}

	lb.setDNSFields(*lbDescription)

	return &lb, nil
}

func (lb ELB) findExisting() (*elb.LoadBalancerDescription, error) {
	resp, err := lb.Client.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{
			aws.String(lb.Name),
		},
		PageSize: aws.Int64(1),
	})

	if err != nil {
		return nil, microerror.Mask(err)
	}

	descriptions := resp.LoadBalancerDescriptions

	if len(descriptions) < 1 {
		return nil, microerror.Maskf(notFoundError, notFoundErrorFormat, ELBType, lb.Name)
	} else if len(descriptions) > 1 {
		return nil, microerror.Mask(tooManyResultsError)
	}

	return descriptions[0], nil
}

func (lb *ELB) setDNSFields(desc elb.LoadBalancerDescription) {
	lb.dnsName = *desc.DNSName
	lb.hostedZoneID = *desc.CanonicalHostedZoneNameID
}
