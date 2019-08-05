package env

import (
	"fmt"
	"os"
)

const (
	// EnvVarCircleCI is the process environment variable representing the
	// CIRCLECI env var.
	EnvVarCircleCI = "CIRCLECI"
	// EnvVarCircleSHA is the process environment variable representing the
	// CIRCLE_SHA1 env var.
	EnvVarCircleSHA = "CIRCLE_SHA1"
	// EnvVarE2EKubeconfig is the process environment variable representing the
	// E2E_KUBECONFIG env var.
	EnvVarE2EKubeconfig = "E2E_KUBECONFIG"
	// EnvVarKeepResources is the process environment variable representing the
	// KEEP_RESOURCES env var.
	EnvVarKeepResources = "KEEP_RESOURCES"

	// e2eHarnessDefaultKubeconfig is defined to avoid dependency of
	// e2e-harness. e2e-harness depends on this project. We don't want
	// circular dependencies even though it works in this case. This makes
	// vendoring very tricky.
	//
	// NOTE this should reflect value of DefaultKubeConfig constant.
	//
	//	See https://godoc.org/github.com/giantswarm/e2e-harness/pkg/harness#pkg-constants.
	//
	// There is also a note in the code there.
	//
	//	See https://github.com/giantswarm/e2e-harness/pull/177
	//
	e2eHarnessDefaultKubeconfig = "/workdir/.shipyard/config"
)

var (
	circleCI      string
	circleSHA     string
	keepResources string
	kubeconfig    string
)

func init() {
	circleCI = os.Getenv(EnvVarCircleCI)
	keepResources = os.Getenv(EnvVarKeepResources)

	kubeconfig = os.Getenv(EnvVarE2EKubeconfig)
	if kubeconfig == "" {
		kubeconfig = e2eHarnessDefaultKubeconfig
	}

	circleSHA = os.Getenv(EnvVarCircleSHA)
	if circleSHA == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarCircleSHA))
	}
}

func CircleCI() string {
	return circleCI
}

func CircleSHA() string {
	return circleSHA
}

func KeepResources() string {
	return keepResources
}

func KubeConfigPath() string {
	return kubeconfig
}
