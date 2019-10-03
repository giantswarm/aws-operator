// +build k8srequired

package env

import (
	"crypto/sha1"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

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

	// IDChars represents the character set used to generate cluster IDs.
	// (does not contain 1 and l, to avoid confusion)
	IDChars = "023456789abcdefghijkmnopqrstuvwxyz"
	// IDLength represents the number of characters used to create a cluster ID.
	IDLength = 3
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

	// ClusterID returns a cluster ID unique to a run integration test. It might
	// look like ci-w3e95.
	//
	//     ci is a static identifier stating a CI run of the aws-operator.
	//     w is a version reference for wip which can also be c for the current version.
	//     3 is the first character of the Git SHA.
	//     e95 is a randomly generated alphanumeric string.
	//
	var parts []string
	parts = append(parts, "ci-")
	parts = append(parts, TestedVersion()[0:1])
	parts = append(parts, CircleSHA()[0:1])
	parts = append(parts, generateID())
	clusterID = strings.Join(parts, "")

}

func CircleSHA() string {
	return circleSHA
}

func ClusterID() string {
	return clusterID
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

// generateId returns a string to be used as unique cluster ID
func generateID() string {
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	for {
		letterRunes := []rune(IDChars)
		b := make([]rune, IDLength)
		for i := range b {
			b[i] = letterRunes[rng.Intn(len(letterRunes))]
		}

		id := string(b)

		if _, err := strconv.Atoi(id); err == nil {
			// string is numbers only, which we want to avoid
			continue
		}

		matched, err := regexp.MatchString("^[a-z]+$", id)
		if err == nil && matched == true {
			// strings is letters only, which we also avoid
			continue
		}

		return id
	}
}
