package cloudconfig

import (
	"context"
)

type Interface interface {
	// NewPaths returns a list of S3 Object paths aligned with the templates
	// returned by NewTemplates.
	NewPaths(ctx context.Context, obj interface{}) ([]string, error)
	// NewTemplates implements any functionality necessary to generate a list of
	// Cloud Config templates. The interface defintion is most generic in order to
	// serve all possible cases. The returned template is a list of Clooud Config
	// templates ready to upload to S3. Usually the amount of templates generated
	// should be 1. There may be special cases though e.g. HA Masters, where an
	// implementation may detect an HA Masters setting and thus needs to generate
	// multiple Cloud Configs based on e.g. some desired replicas configuration.
	// Just like NewPaths, the implementation of NewTemplates must align with the
	// returned items so that users of the interface are guaranteed to always work
	// with a key-value pair of path and template.
	NewTemplates(ctx context.Context, obj interface{}) ([]string, error)
}
