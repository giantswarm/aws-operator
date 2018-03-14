// +build k8srequired

package env

import (
	"crypto/sha1"
	"fmt"
	"os"
	"strings"

	"github.com/giantswarm/e2e-harness/pkg/framework"

	"github.com/giantswarm/aws-operator/service"
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
	// EnvVarTestDir is the process environment variable representing the
	// TEST_DIR env var.
	EnvVarTestDir = "TEST_DIR"
	// EnvVarVersionBundleVersion is the process environment variable representing
	// the VERSION_BUNDLE_VERSION env var.
	EnvVarVersionBundleVersion = "VERSION_BUNDLE_VERSION"
)

var (
	circleSHA            string
	testedVersion        string
	testDir              string
	versionBundleVersion string
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

	testDir = os.Getenv(EnvVarTestDir)

	// NOTE that implications of changing the order of initialization here means
	// breaking the initialization behaviour.
	clusterID := os.Getenv(EnvVarClusterID)
	if clusterID == "" {
		os.Setenv(EnvVarClusterID, ClusterID())
	}

	var err error
	versionBundleVersion, err = framework.GetVersionBundleVersion(service.NewVersionBundles(), TestedVersion())
	if err != nil {
		panic(err.Error())
	}
	// TODO there should be a not found error returned by the framework in such
	// cases.
	if VersionBundleVersion() == "" {
		panic("version bundle version  must not be empty")
	}
	os.Setenv(EnvVarVersionBundleVersion, VersionBundleVersion())
}

func CircleSHA() string {
	return circleSHA
}

// ClusterID returns a cluster ID unique to a run integration test. It might
// look like ci-wip-3cc75-5e958.
//
//     ci is a static identifier stating a CI run of the aws-operator.
//     wip is a version reference which can also be cur for the current version.
//     3cc75 is the Git SHA.
//     5e958 is a hash of the integration test dir, if any.
//
func ClusterID() string {
	var parts []string

	parts = append(parts, "ci")
	parts = append(parts, TestedVersion()[0:3])
	parts = append(parts, CircleSHA()[0:5])
	if TestHash() != "" {
		parts = append(parts, TestHash())
	}

	return strings.Join(parts, "-")
}

func TestedVersion() string {
	return testedVersion
}

func TestDir() string {
	return testDir
}

func TestHash() string {
	if TestDir() == "" {
		return ""
	}

	h := sha1.New()
	h.Write([]byte(TestDir()))
	s := fmt.Sprintf("%x", h.Sum(nil))[0:5]

	return s
}

func VersionBundleVersion() string {
	return versionBundleVersion
}
