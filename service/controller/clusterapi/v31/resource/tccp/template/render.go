package template

import (
	"fmt"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/pkg/template"
)

func Render(v interface{}) (string, error) {
	l := []string{
		TemplateMain,
		TemplateMainIAMPolicies,
		TemplateMainInstance,
		TemplateMainInternetGateway,
		TemplateMainLoadBalancers,
		TemplateMainNatGateway,
		TemplateMainOutputs,
		TemplateMainRecordSets,
		TemplateMainRouteTables,
		TemplateMainSecurityGroups,
		TemplateMainSubnets,
		TemplateMainVPC,
	}

	s, err := template.Render(l, v)
	if err != nil {
		return "", microerror.Mask(err)
	}

	fmt.Printf("%s\n", s)

	return s, nil
}
