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
	component = "aws-operator"
	provider  = "aws"
)

const (
	EnvVarCircleCI             = "CIRCLECI"
	EnvVarCircleSHA            = "CIRCLE_SHA1"
	EnvVarE2EKubeconfig        = "E2E_KUBECONFIG"
	EnvVarGithubBotToken       = "GITHUB_BOT_TOKEN"
	EnvVarKeepResources        = "KEEP_RESOURCES"
	EnvVarRegistryPullSecret   = "REGISTRY_PULL_SECRET"
	EnvVarTestedVersion        = "E2E_TESTED_VERSION"
	EnvVarTestDir              = "E2E_TEST_DIR"
	EnvVarVersionBundleVersion = "VERSION_BUNDLE_VERSION"
)

var (
	circleCI             string
	circleSHA            string
	clusterID            string
	kubeconfigPath       string
	registryPullSecret   string
	githubToken          string
	testDir              string
	testedVersion        string
	keepResources        string
	versionBundleVersion string
)

func init() {
	var err error

	circleCI = os.Getenv(EnvVarCircleCI)
	keepResources = os.Getenv(EnvVarKeepResources)

	circleSHA = os.Getenv(EnvVarCircleSHA)
	if circleSHA == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarCircleSHA))
	}

	githubToken = os.Getenv(EnvVarGithubBotToken)
	if githubToken == "" {
		panic(fmt.Sprintf("env var %q must not be empty", EnvVarGithubBotToken))
	}

	kubeconfigPath = os.Getenv(EnvVarE2EKubeconfig)
	if kubeconfigPath == "" {
		panic(fmt.Sprintf("env var %q must not be empty", EnvVarE2EKubeconfig))
	}

	testedVersion = os.Getenv(EnvVarTestedVersion)
	if testedVersion == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarTestedVersion))
	}

	registryPullSecret = os.Getenv(EnvVarRegistryPullSecret)
	if registryPullSecret == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarRegistryPullSecret))
	}

	testDir = os.Getenv(EnvVarTestDir)

	params := &framework.VBVParams{
		Component: component,
		Provider:  provider,
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
	parts = append(parts, TestedVersion()[0:1])
	parts = append(parts, CircleSHA()[0:2])
	if TestHash() != "" {
		parts = append(parts, TestHash())
	}

	return strings.Join(parts, "")
}

func KeepResources() bool {
	return keepResources == strings.ToLower("true")
}

func GithubToken() string {
	return githubToken
}

func KubeConfigPath() string {
	return kubeconfigPath
}

func RegistryPullSecret() string {
	return registryPullSecret
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
