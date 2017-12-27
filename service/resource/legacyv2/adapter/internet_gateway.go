package adapter

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/keyv2"
)

// template related to this adapter: service/templates/cloudformation/internet_gateway.yaml

type internetGatewayAdapter struct {
	VPCID string
}

func (i *internetGatewayAdapter) getInternetGateway(customObject v1alpha1.AWSConfig, clients Clients) error {
	// TODO: remove this code once the VPC is created by cloudformation and add a
	// reference in the template
	vpcID, err := VPCID(clients, keyv2.ClusterID(customObject))
	if err != nil {
		return microerror.Mask(err)
	}
	i.VPCID = vpcID

	return nil
}
