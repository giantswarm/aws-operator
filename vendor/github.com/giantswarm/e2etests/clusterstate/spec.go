package clusterstate

import (
	"context"
)

const (
	CNRAddress      = "https://quay.io"
	CNROrganization = "giantswarm"
	ChartChannel    = "stable"
	ChartName       = "e2e-app-chart"
	ChartNamespace  = "e2e-app"
)

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
