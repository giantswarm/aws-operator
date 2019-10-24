package env

import (
	"crypto/sha1"
	"fmt"
	"math/rand"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/giantswarm/aws-operator/service"
	v28 "github.com/giantswarm/aws-operator/service/controller/legacy/v28"
)

const (
	component = "aws-operator"
	provider  = "aws"
)

const (
	EnvVarCircleCI           = "CIRCLECI"
	EnvVarCircleSHA          = "CIRCLE_SHA1"
	EnvVarGithubBotToken     = "GITHUB_BOT_TOKEN"
	EnvVarKeepResources      = "KEEP_RESOURCES"
	EnvVarRegistryPullSecret = "REGISTRY_PULL_SECRET"
	EnvVarTestDir            = "TEST_DIR"

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
	registryPullSecret   string
	githubToken          string
	testDir              string
	keepResources        string
	versionBundleVersion string
)

func init() {
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

	registryPullSecret = os.Getenv(EnvVarRegistryPullSecret)
	if registryPullSecret == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarRegistryPullSecret))
	}

	testDir = os.Getenv(EnvVarTestDir)

	{
		vbs := service.NewVersionBundles()

		if path.Base(testDir) == "update" {
			// For the update test we want to create previous
			// version so we can upgrade from it.

			// Versions v29patch1 and v29 are broken can it is not
			// possible to upgrade from them so fixed version for
			// bundle v28 is returned. If they were ok the code
			// blow should look like:
			//
			//	versionBundleVersion = vbs[len(vbs)-2].Version
			//
			versionBundleVersion = v28.VersionBundle().Version
		} else {
			versionBundleVersion = vbs[len(vbs)-1].Version
		}
	}

	// init clusterID
	rand.Seed(time.Now().UnixNano())
	var parts []string
	parts = append(parts, "ci-")
	parts = append(parts, CircleSHA()[0:2])
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

		matched, err := regexp.MatchString("^[a-z]+$", id)
		if err == nil && matched == true {
			// strings is letters only, which we also avoid
			continue
		}

		return id
	}
}
