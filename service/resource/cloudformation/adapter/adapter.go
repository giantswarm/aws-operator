// Package adapter contains the required logic for creating data structures used for
// feeding CloudFormation templates.
//
// It follows the adapter pattern https://en.wikipedia.org/wiki/Adapter_pattern in the
// sense that has the knowledge to transform a aws custom object into a data structure
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
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/microerror"
)

type hydrater func(awstpr.CustomObject, Clients) error

type Adapter struct {
	ASGType string

	launchConfigAdapter
	autoScalingGroupAdapter
}

func New(customObject awstpr.CustomObject, clients Clients) (Adapter, error) {
	a := Adapter{}

	a.ASGType = prefixWorker

	hydraters := []hydrater{
		a.getAutoScalingGroup,
		a.getLaunchConfiguration,
	}

	for _, h := range hydraters {
		if err := h(customObject, clients); err != nil {
			return Adapter{}, microerror.Mask(err)
		}
	}

	return a, nil
}
