// +build k8srequired

package integration

import (
	"fmt"
	"os"
)

const (
	// EnvVarTestedVersion is the process environment variable representing the
	// TESTED_VERSION env var.
	EnvVarTestedVersion = "TESTED_VERSION"
)

var (
	testedVersion string
)

func init() {
	testedVersion = os.Getenv(EnvVarTestedVersion)
	if testedVersion == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarTestedVersion))
	}
}

func TestedVersion() string {
	return testedVersion
}
