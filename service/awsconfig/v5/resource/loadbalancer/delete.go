package loadbalancer

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	deleteInput, err := toLoadBalancerState(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if deleteInput != nil && len(deleteInput.LoadBalancerNames) > 0 {
		for _, lbName := range deleteInput.LoadBalancerNames {
			_, err := r.clients.ELB.DeleteLoadBalancer(&elb.DeleteLoadBalancerInput{
				LoadBalancerName: aws.String(lbName),
			})
			if err != nil {
				return microerror.Mask(err)
			}
		}
	}

	r.logger.LogCtx(ctx, "debug", fmt.Sprintf("deleted %d load balancers", len(deleteInput.LoadBalancerNames)))

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	delete, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
	patch.SetDeleteChange(delete)

	return patch, nil
}

func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentLBState, err := toLoadBalancerState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredLBState, err := toLoadBalancerState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var lbStateToDelete *LoadBalancerState
	if desiredLBState == nil && currentLBState != nil && len(currentLBState.LoadBalancerNames) > 0 {
		lbStateToDelete = currentLBState
	}

	return lbStateToDelete, nil
}
