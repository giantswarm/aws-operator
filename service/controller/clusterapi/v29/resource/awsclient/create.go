package awsclient

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := r.toClusterFunc(obj)
	fmt.Printf("\n")
	fmt.Printf("\n")
	fmt.Printf("\n")
	fmt.Printf("    err:  %#v\n", microerror.Cause(err))
	fmt.Printf("    ReasonForError(err): %#v\n", errors.ReasonForError(microerror.Cause(err)))
	fmt.Printf("    metav1.StatusReasonNotFound: %#v\n", metav1.StatusReasonNotFound)
	fmt.Printf("\n")
	fmt.Printf("\n")
	fmt.Printf("\n")
	if IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "cluster cr not yet availabile")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	err = r.addAWSClientsToContext(ctx, cr)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
