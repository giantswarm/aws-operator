package env

import (
	"crypto/sha1" //nolint:gosec
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/microerror"
)

const (
	component = "aws-operator"
	provider  = "aws"
)

const (
	EnvVarCircleSHA            = "CIRCLE_SHA1"
	EnvVarGithubBotToken       = "GITHUB_BOT_TOKEN" //nolint:gosec
	EnvVarKeepResources        = "KEEP_RESOURCES"
	EnvVarRegistryPullSecret   = "REGISTRY_PULL_SECRET" //nolint:gosec
	EnvVarTestedVersion        = "TESTED_VERSION"
	EnvVarTestDir              = "TEST_DIR"
	EnvVarVersionBundleVersion = "VERSION_BUNDLE_VERSION"

	// IDChars represents the character set used to generate cluster IDs.
	// (does not contain 1 and l, to avoid confusion)
	IDChars = "023456789abcdefghijkmnopqrstuvwxyz"
	// IDLength represents the number of characters used to create a cluster ID.
	IDLength = 3
)

var (
	circleSHA            string
	clusterID            string
	registryPullSecret   string
	githubToken          string
	testDir              string
	testedVersion        string
	keepResources        string
	versionBundleVersion string
)

func init() {
	var err error

	keepResources = os.Getenv(EnvVarKeepResources)

	circleSHA = os.Getenv(EnvVarCircleSHA)
	if circleSHA == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarCircleSHA))
	}

	githubToken = os.Getenv(EnvVarGithubBotToken)
	if githubToken == "" {
		panic(fmt.Sprintf("env var %q must not be empty", EnvVarGithubBotToken))
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

	// init clusterID
	rand.Seed(time.Now().UnixNano())
	var parts []string
	parts = append(parts, "ci-")
	parts = append(parts, TestedVersion()[0:1])
	parts = append(parts, CircleSHA()[0:1])
	parts = append(parts, generateID(IDLength))
	clusterID = strings.Join(parts, "")
}

func CircleSHA() string {
	return circleSHA
}

// ClusterID returns a cluster ID unique to a run integration test. It might
// look like ci-w3e95.
//
//     ci is a static identifier stating a CI run of the aws-operator.
//     w is a version reference which can also be c for the current version.
//     3 is the first character of the Git SHA.
//     e95 is a randomly generated alphanumeric string.
//
func ClusterID() string {
	return clusterID
}

func KeepResources() bool {
	return keepResources == strings.ToLower("true")
}

func GithubToken() string {
	return githubToken
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

	h := sha1.New() //nolint:gosec
	_, err := h.Write([]byte(TestDir()))
	if err != nil {
		panic(microerror.JSON(err))
	}

	s := fmt.Sprintf("%x", h.Sum(nil))[0:5]

	return s
}

func VersionBundleVersion() string {
	return versionBundleVersion
}

// generateID returns a string to be used as unique cluster ID
func generateID(idLength int) string {
	for {
		letterRunes := []rune(IDChars)
		b := make([]rune, idLength)
		for i := range b {
			b[i] = letterRunes[rand.Intn(len(letterRunes))]
		}

		id := string(b)

		if _, err := strconv.Atoi(id); err == nil {
			// string is numbers only, which we want to avoid
			continue
		}

		matched, err := regexp.MatchString("^[a-z]+$", id) //nolint:staticcheck
		if err == nil && matched {
			// strings is letters only, which we also avoid
			continue
		}

		return id
	}
}
