package clusterstate

import (
	"context"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	CNRAddress      = "https://quay.io"
	CNROrganization = "giantswarm"
	ChartChannel    = "stable"
	ChartName       = "e2e-app-chart"
	ChartNamespace  = "e2e-app"
)

type LegacyFramework interface {
	// K8sClient returns a properly configured tenant cluster client for the
	// Kubernetes API.
	K8sClient() kubernetes.Interface
	// RestConfig returns the rest config used to generate the Kubernetes client as
	// returned by K8sClient.
	RestConfig() *rest.Config
	// WaitForAPIUp waits for the currently configured tenant cluster Kubernetes
	// API to be down.
	WaitForAPIDown() error
	// WaitForGuestReady waits for the currently configured tenant cluster to be
	// ready.
	WaitForGuestReady(ctx context.Context) error
}

type Interface interface {
	// Test executes the cluster state test using the configured provider
	// implementation. The provider implementation has to be aware of the guest
	// cluster it has to act against. The test processes the following steps to
	// ensure the cluster state persists when rebooting and replacing the
	// master node.
	//
	//  - Install test app.
	//  - Check test app is installed.
	//  - Reboot master node.
	//  - Wait for API to be down.
	//  - Wait for cluster to be ready.
	//  - Check test app is installed.
	//  - Replace master node.
	//  - Wait for API to be down.
	//  - Wait for cluster to be ready.
	//  - Check test app is installed.
	//
	Test(ctx context.Context) error
}
