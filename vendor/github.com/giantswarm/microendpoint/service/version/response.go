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
	VersionBundles []versionbundle.Bundle `json:"version_bundles"`
}

// DefaultResponse provides a default response object by best effort.
func DefaultResponse() *Response {
	return &Response{
		Description:    "",
		GitCommit:      "",
		GoVersion:      "",
		Name:           "",
		OSArch:         "",
		Source:         "",
		VersionBundles: nil,
	}
}
