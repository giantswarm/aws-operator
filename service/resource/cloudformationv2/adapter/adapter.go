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
// definition and the hydrater function to fill the fields (like asg.go or
// launch_configuration.go).
// * Add the new type to the Adapter type in `service/resource/cloudformation/adapter/adapter.go`
// and include the hydrater function in the `hydraters` slice.
package adapter

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/keyv2"
)

type hydrater func(Config) error

type Adapter struct {
	ASGType          string
	AvailabilityZone string
	ClusterID        string
	MasterImageID    string
	WorkerImageID    string

	autoScalingGroupAdapter
	iamPoliciesAdapter
	hostIamRolesAdapter
	instanceAdapter
	launchConfigAdapter
	loadBalancersAdapter
	recordSetsAdapter
	routeTablesAdapter
	securityGroupsAdapter
	hostRouteTablesAdapter
	subnetsAdapter
	vpcAdapter
	outputsAdapter
}

type Config struct {
	CustomObject     v1alpha1.AWSConfig
	Clients          Clients
	HostClients      Clients
	InstallationName string
	HostAccountID    string
	GuestAccountID   string
}

func NewGuest(cfg Config) (Adapter, error) {
	a := Adapter{}

	a.ASGType = prefixWorker
	a.ClusterID = keyv2.ClusterID(cfg.CustomObject)
	a.MasterImageID = keyv2.MasterImageID(cfg.CustomObject)
	a.WorkerImageID = keyv2.WorkerImageID(cfg.CustomObject)

	hydraters := []hydrater{
		a.getAutoScalingGroup,
		a.getIamPolicies,
		a.getInstance,
		a.getLaunchConfiguration,
		a.getLoadBalancers,
		a.getRecordSets,
		a.getRouteTables,
		a.getSecurityGroups,
		a.getSubnets,
		a.getVpc,
		a.getOutputs,
	}

	for _, h := range hydraters {
		if err := h(cfg); err != nil {
			return Adapter{}, microerror.Mask(err)
		}
	}

	return a, nil
}

func NewHostPre(cfg Config) (Adapter, error) {
	a := Adapter{}

	hydraters := []hydrater{
		a.getHostIamRoles,
	}

	for _, h := range hydraters {
		if err := h(cfg); err != nil {
			return Adapter{}, microerror.Mask(err)
		}
	}

	return a, nil
}

func NewHostPost(cfg Config) (Adapter, error) {
	a := Adapter{}

	hydraters := []hydrater{
		a.getHostRouteTables,
	}

	for _, h := range hydraters {
		if err := h(cfg); err != nil {
			return Adapter{}, microerror.Mask(err)
		}
	}

	return a, nil
}
