// Package adapter contains the required logic for creating data structures used for
// feeding CloudFormation templates.
//
// It follows the adapter pattern https://en.wikipedia.org/wiki/Adapter_pattern in the
// sense that it has the knowledge to transform a aws custom object into a data structure
// easily interpolable into the templates without any additional view logic.
//
// There's a base template in `service/templates/cloudformation/main.yaml` which defines
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
// * Add the template file in `service/template/cloudformation` and modify
// `service/template/cloudformation/main.yaml` to include the new template.
// * Add the adapter logic file in `service/resource/cloudformation/adapter` with the type
// definition and the hydrater function to fill the fields (like asg.go or
// launch_configuration.go).
// * Add the new type to the Adapter type in `service/resource/cloudformation/adapter/adapter.go`
// and include the hydrater function in the `hydraters` slice.
package adapter

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/aws-operator/service/keyv2"
	"github.com/giantswarm/microerror"
)

type hydrater func(v1alpha1.AWSConfig, Clients) error

type Adapter struct {
	ASGType          string
	AvailabilityZone string
	ClusterID        string

	autoScalingGroupAdapter
	instanceAdapter
	launchConfigAdapter
	loadBalancersAdapter
	outputsAdapter
	recordSetsAdapter
	workerPolicyAdapter
}

func New(customObject v1alpha1.AWSConfig, clients Clients) (Adapter, error) {
	a := Adapter{}

	a.ASGType = prefixWorker
	a.ClusterID = keyv2.ClusterID(customObject)

	hydraters := []hydrater{
		a.getAutoScalingGroup,
		a.getInstance,
		a.getLaunchConfiguration,
		a.getLaunchConfiguration,
		a.getLoadBalancers,
		a.getOutputs,
		a.getRecordSets,
		a.getWorkerPolicy,
	}

	for _, h := range hydraters {
		if err := h(customObject, clients); err != nil {
			return Adapter{}, microerror.Mask(err)
		}
	}

	return a, nil
}
