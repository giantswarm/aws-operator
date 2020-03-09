package controllercontext

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/aws-operator/client/aws"
)

type ContextClient struct {
	ControlPlane  ContextClientControlPlane
	TenantCluster ContextClientTenantCluster
}

type ContextClientControlPlane struct {
	AWS aws.Clients
}

type ContextClientTenantCluster struct {
	AWS aws.Clients
	G8s versioned.Interface
	K8s kubernetes.Interface
}
