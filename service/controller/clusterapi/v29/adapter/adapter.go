// Package adapter contains the required logic for creating data structures used for
// feeding CloudFormation templates.
//
// It follows the adapter pattern https://en.wikipedia.org/wiki/Adapter_pattern in the
// sense that it has the knowledge to transform a aws custom object into a data structure
// easily interpolable into the templates without any additional view logic.
//
// There's a base template in `service/templates/cloudformation/guest/main.yaml` which defines
// the basic structure and includes the rest of templates that form the stack as nested
// templates. Those subtemplates should use a `define` action with the name that will be
// used to refer to them from the main template, as explained here
// https://golang.org/pkg/text/template/#hdr-Nested_template_definitions
//
// Each adapter is related to one of these nested templates. It includes the data structure
// with all the values needed to interpolate in the related template and the logic required
// to obtain them, this logic is packed into functions called `hydraters`.
//
// When extending the stack we will just need to:
// * Add the template file in `service/template/cloudformation/guest` and modify
// `service/template/cloudformation/main.yaml` to include the new template.
// * Add the adapter logic file in `service/resource/cloudformation/adapter` with the type
// definition and the Hydrater function to fill the fields (like asg.go or
// launch_configuration.go).
// * Add the new type to the Adapter type in `service/resource/cloudformation/adapter/adapter.go`
// and include the Hydrater function in the `hydraters` slice.
package adapter

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

type Config struct {
	APIWhitelist                    APIWhitelist
	ControlPlaneAccountID           string
	ControlPlaneNATGatewayAddresses []*ec2.Address
	ControlPlanePeerRoleARN         string
	ControlPlaneVPCID               string
	ControlPlaneVPCCidr             string
	CustomObject                    v1alpha1.Cluster
	EncrypterBackend                string
	GuestAccountID                  string
	InstallationName                string
	MachineDeployment               v1alpha1.MachineDeployment
	PublicRouteTables               string
	Route53Enabled                  bool
	StackState                      StackState
	TenantClusterAccountID          string
	TenantClusterKMSKeyARN          string
}

type Adapter struct {
	Guest GuestAdapter
}

func NewGuest(cfg Config) (Adapter, error) {
	a := Adapter{}

	hydraters := []Hydrater{
		a.Guest.AutoScalingGroup.Adapt,
		a.Guest.IAMPolicies.Adapt,
		a.Guest.InternetGateway.Adapt,
		a.Guest.Instance.Adapt,
		a.Guest.LaunchConfiguration.Adapt,
		a.Guest.LifecycleHooks.Adapt,
		a.Guest.LoadBalancers.Adapt,
		a.Guest.NATGateway.Adapt,
		a.Guest.Outputs.Adapt,
		a.Guest.RecordSets.Adapt,
		a.Guest.RouteTables.Adapt,
		a.Guest.SecurityGroups.Adapt,
		a.Guest.Subnets.Adapt,
		a.Guest.VPC.Adapt,
	}

	for _, h := range hydraters {
		if err := h(cfg); err != nil {
			return Adapter{}, microerror.Mask(err)
		}
	}

	return a, nil
}

type GuestAdapter struct {
	AutoScalingGroup    GuestAutoScalingGroupAdapter
	IAMPolicies         GuestIAMPoliciesAdapter
	InternetGateway     GuestInternetGatewayAdapter
	Instance            GuestInstanceAdapter
	LaunchConfiguration GuestLaunchConfigAdapter
	LifecycleHooks      GuestLifecycleHooksAdapter
	LoadBalancers       GuestLoadBalancersAdapter
	NATGateway          GuestNATGatewayAdapter
	Outputs             GuestOutputsAdapter
	RecordSets          GuestRecordSetsAdapter
	RouteTables         GuestRouteTablesAdapter
	SecurityGroups      GuestSecurityGroupsAdapter
	Subnets             GuestSubnetsAdapter
	VPC                 GuestVPCAdapter
}
