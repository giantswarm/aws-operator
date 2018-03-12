// +build k8srequired

package integration

import (
	"fmt"
	"os"
)

const (
	// EnvVarCircleSHA is the process environment variable representing the
	// CIRCLE_SHA1 env var.
	EnvVarCircleSHA = "CIRCLE_SHA1"
	// EnvVarClusterID is the process environment variable representing the
	// CLUSTER_NAME env var.
	//
	// TODO rename to CLUSTER_ID. Note this also had to be changed in the
	// framework package of e2e-harness.
	EnvVarClusterID = "CLUSTER_NAME"
	// EnvVarTestedVersion is the process environment variable representing the
	// TESTED_VERSION env var.
	EnvVarTestedVersion = "TESTED_VERSION"
)

var (
	circleSHA     string
	testedVersion string
)

func init() {
	circleSHA = os.Getenv(EnvVarCircleSHA)
	if circleSHA == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarCircleSHA))
	}

	testedVersion = os.Getenv(EnvVarTestedVersion)
	if testedVersion == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarTestedVersion))
	}

	// NOTE that implications of changing the order of initialization here means
	// breaking the initialization behaviour.
	clusterID := os.Getenv(EnvVarClusterID)
	if clusterID == "" {
		os.Setenv(EnvVarClusterID, ClusterID())
	}
}

func CircleSHA() string {
	return circleSHA
}

func ClusterID() string {
	return fmt.Sprintf("ci-awsop-%s-%s", TestedVersion(), CircleSHA()[0:5])
}

func TestedVersion() string {
	return testedVersion
}
