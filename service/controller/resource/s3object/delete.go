package s3object

import (
	"context"

	"github.com/giantswarm/operatorkit/resource/crud"
)

// ApplyDeleteChange is a noop as deletion for now is covered with the deletion
// of the whole Tenant Cluster. That way all S3 Objects will vanish with the S3
// Bucket. Note that we share the resource implementation to cover different
// Cloud Config objects, for instance for TCCP and TCNP stacks. These have
// different lifecycles which means we do not delete Cloud Config objects of a
// deleted Node Pool. Another very rare but noteworthy side effect might be that
// Node Pool IDs generate twice during the lifetime of a Tenant Cluster cause
// existing Cloud Config objects to be overwritten with newer versions,
// potentially causing confusion and inconsistencies upon inspection. Chances are
// you win the lotto before this bullshit ever happens, but one should
// understand current design decisions.
func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	return nil
}

// NewDeletePatch is a noop like ApplyDeleteChange. See the godoc there.
func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*crud.Patch, error) {
	return nil, nil
}
