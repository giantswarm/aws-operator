package k8sclient

const (
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
