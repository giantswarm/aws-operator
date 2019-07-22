package version

import "github.com/giantswarm/versionbundle"

// Response is the return value of the service action.
type Response struct {
	Description    string                 `json:"description"`
	GitCommit      string                 `json:"git_commit"`
	GoVersion      string                 `json:"go_version"`
	Name           string                 `json:"name"`
	OSArch         string                 `json:"os_arch"`
	Source         string                 `json:"source"`
	Version        string                 `json:"version"`
	VersionBundles []versionbundle.Bundle `json:"version_bundles"`
}
