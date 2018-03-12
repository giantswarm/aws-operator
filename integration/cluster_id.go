// +build k8srequired

package integration

import (
	"fmt"
	"os"
)

const (
	// EnvVarClusterID is the process environment variable representing the
	// CLUSTER_NAME env var.
	//
	// TODO rename to CLUSTER_ID. Note this also had to be changed in the
	// framework package of e2e-harness.
	EnvVarClusterID = "CLUSTER_NAME"
)

// init sets the cluster ID of the current integration test to the process
// environment at runtime. init of cluster_id.go is called after init of
// circle_sha.go. Execution of multiple init functions within a package is
// guaranteed to happen in lexicographical order.
//
//     To ensure reproducible initialization behavior, build systems are
//     encouraged to present multiple files belonging to the same package in
//     lexical file name order to a compiler.
//
// NOTE Implications of changing file names means breaking the initialization
// behaviour.
func init() {
	os.Setenv(EnvVarClusterID, ClusterID())
}

func ClusterID() string {
	return fmt.Sprintf("ci-awsop-%s", CircleSHA()[0:5])
}
