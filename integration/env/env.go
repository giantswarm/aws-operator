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

var (
	awsAPIHostedZoneGuest     string
	awsIngressHostedZoneGuest string
	awsRegion                 string
	circleSHA                 string
	clusterName               string
	commonDomain              string
	guestAWSARN               string
	guestAWSAccessKeyID       string
	guestAWSAccessKeySecret   string
	guestAWSAccessKeyToken    string
	hostAWSAccessKeyID        string
	hostAWSAccessKeySecret    string
	hostAWSAccessKeyToken     string
	idRSAPub                  string
	testedVersion             string
	testDir                   string
	registryPullSecret        string
	vaultToken                string
	versionBundleVersion      string
)

func init() {
	getenv := func(key string, value *string) {
		*value = os.Getenv(key)
		if *value == "" {
			panic(fmt.Sprintf("env var '%s' must not be empty", key))
		}
	}

	var blackHole string

	getenv("AWS_API_HOSTED_ZONE_GUEST", &awsAPIHostedZoneGuest)
	getenv("AWS_INGRESS_HOSTED_ZONE_GUEST", &awsIngressHostedZoneGuest)
	getenv("AWS_REGION", &awsRegion)
	getenv("CIRCLE_SHA1", &circleSHA)
	// TODO rename to CLUSTER_ID. Note this also had to be changed in the
	// framework package of e2e-harness.
	getenv("CLUSTER_NAME", &clusterName)
	getenv("COMMON_DOMAIN", &commonDomain)
	getenv("GITHUB_BOT_TOKEN", &blackHole)
	getenv("GUEST_AWS_ARN", &guestAWSARN)
	getenv("GUEST_AWS_ACCESS_KEY_ID", &guestAWSAccessKeyID)
	getenv("GUEST_AWS_SECRET_ACCESS_KEY", &guestAWSAccessKeySecret)
	getenv("GUEST_AWS_SESSION_TOKEN", &guestAWSAccessKeyToken)
	getenv("HOST_AWS_ACCESS_KEY_ID", &hostAWSAccessKeyID)
	getenv("HOST_AWS_SECRET_ACCESS_KEY", &hostAWSAccessKeySecret)
	getenv("HOST_AWS_SESSION_TOKEN", &hostAWSAccessKeyToken)
	getenv("IDRSA_PUB", &idRSAPub)
	getenv("TESTED_VERSION", &testedVersion)
	getenv("TEST_DIR", &testDir)
	getenv("REGISTRY_PULL_SECRET", &registryPullSecret)
	getenv("VAULT_TOKEN", &vaultToken)

	// This get versionBundleVersion. This should be remove with release-operator.
	{
		var err error
		var token string
		getenv("GITHUB_BOT_TOKEN", &token)

		params := &framework.VBVParams{
			Component: "aws-operator",
			Provider:  "aws",
			Token:     token,
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
	}
}

func AWSAPIHostedZoneGuest() string {
	return awsAPIHostedZoneGuest
}

func AWSIngressHostedZoneGuest() string {
	return awsIngressHostedZoneGuest
}

func AWSRegion() string {
	return awsRegion
}

func AWSRouteTable0() string {
	return ClusterID() + "_0"
}

func AWSRouteTable1() string {
	return ClusterID() + "_1"
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

func ClusterName() string {
	return clusterName
}

func CommonDomain() string {
	return commonDomain
}

func GuestAWSARN() string {
	return guestAWSARN
}

func GuestAWSAccessKeyID() string {
	return guestAWSAccessKeyID
}

func GuestAWSAccessKeySecret() string {
	return guestAWSAccessKeySecret
}

func GuestAWSAccessKeyToken() string {
	return guestAWSAccessKeyToken
}

func HostAWSAccessKeyID() string {
	return hostAWSAccessKeyID
}

func HostAWSAccessKeySecret() string {
	return hostAWSAccessKeySecret
}

func HostAWSAccessKeyToken() string {
	return hostAWSAccessKeyToken
}

func IDRSAPub() string {
	return idRSAPub
}

func TestedVersion() string {
	return testedVersion
}

func TestDir() string {
	return testDir
}

func RegistryPullSecret() string {
	return registryPullSecret
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
