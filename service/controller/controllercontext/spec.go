package controllercontext

import (
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
)

type ContextSpec struct {
	TenantCluster ContextSpecTenantCluster
}

type ContextSpecTenantCluster struct {
	Release releasev1alpha1.Release

	MasterInstance ContextSpecTenantClusterInstance
	WorkerInstance ContextSpecTenantClusterInstance
}

type ContextSpecTenantClusterInstance struct {
	IgnitionHash string
}
