// +build k8srequired

package env

import (
	"crypto/sha1"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/giantswarm/e2e-harness/pkg/framework"
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
	// EnvVarCommonDomain is the process environment variable representing the
	// COMMON_DOMAIN env var.
	EnvVarCommonDomain = "COMMON_DOMAIN"
	// EnvVarGithubBotToken is the process environment variable representing
	// the GITHUB_BOT_TOKEN env var.
	EnvVarGithubBotToken = "GITHUB_BOT_TOKEN"
	// EnvVarGuestAWSArn is the process environment variable representing
	// the GUEST_AWS_ARN env var.
	EnvVarGuestAWSArn = "GUEST_AWS_ARN"
	// EnvVarTestedVersion is the process environment variable representing the
	// TESTED_VERSION env var.
	EnvVarTestedVersion = "TESTED_VERSION"
	// EnvVarTestDir is the process environment variable representing the
	// TEST_DIR env var.
	EnvVarTestDir = "TEST_DIR"
	// EnvVaultToken is the process environment variable representing the
	// VAULT_TOKEN env var.
	EnvVaultToken = "VAULT_TOKEN"
	// EnvVarVersionBundleVersion is the process environment variable representing
	// the VERSION_BUNDLE_VERSION env var.
	EnvVarVersionBundleVersion = "VERSION_BUNDLE_VERSION"
)

var (
	circleSHA            string
	commonDomain         string
	githubToken          string
	guestAWSArn          string
	testedVersion        string
	testDir              string
	vaultToken           string
	versionBundleVersion string
)

func init() {
	circleSHA = os.Getenv(EnvVarCircleSHA)
	if circleSHA == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarCircleSHA))
	}

	guestAWSArn = os.Getenv(EnvVarGuestAWSArn)
	if guestAWSArn == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarGuestAWSArn))
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

	commonDomain = os.Getenv(EnvVarCommonDomain)
	if commonDomain == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarCommonDomain))
	}

	vaultToken = os.Getenv(EnvVaultToken)
	if vaultToken == "" {
		panic(fmt.Sprintf("env var %q must not be empty", EnvVaultToken))
	}

	githubToken = os.Getenv(EnvVarGithubBotToken)
	if githubToken == "" {
		panic(fmt.Sprintf("env var %q must not be empty", EnvVarGithubBotToken))
	}

	var err error
	params := &framework.VBVParams{
		Component: "aws-operator",
		Provider:  "aws",
		Token:     githubToken,
		VType:     TestedVersion(),
	}
	versionBundleVersion, err = framework.GetVersionBundleVersion(params)
	if err != nil {
		panic(err.Error())
	}
	// TODO there should be a not found error returned by the framework in such
	// cases.
	if VersionBundleVersion() == "" {
		if strings.ToLower(TestedVersion()) == "wip" {
			log.Println("WIP version bundle version not present, exiting.")
			os.Exit(0)
		}
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

func CommonDomain() string {
	return commonDomain
}

func GithubToken() string {
	return githubToken
}

func GuestAWSArn() string {
	return guestAWSArn
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

func VaultToken() string {
	return vaultToken
}

func VersionBundleVersion() string {
	return versionBundleVersion
}
