package nodeautorepair

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}
	if !r.enabled {
		r.logger.LogCtx(ctx, "level", "debug", "message", "node auto repair feature is disabled")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil
	}

	if cc.Client.TenantCluster.K8s == nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", "kubernetes clients are not available in controller context yet")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil
	}

	var nodeList corev1.NodeList
	{
		err := cc.Client.TenantCluster.K8s.CtrlClient().List(ctx, &nodeList)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	nodesToTerminate, err := r.detectBadNodes(ctx, nodeList.Items)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d nodes marked for termination", len(nodesToTerminate)))

	maxNodeTermination := maximumNodeTermination(len(nodeList.Items))

	if len(nodesToTerminate) > maxNodeTermination {
		nodesToTerminate = nodesToTerminate[:maxNodeTermination]
	}

	for _, n := range nodesToTerminate {
		err := r.terminateNode(ctx, n, key.ClusterID(&cr))
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (r *Resource) terminateNode(ctx context.Context, node corev1.Node, clusterID string) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var instanceID string
	{
		// node.spec.providerID for AWS is in format aws:///AVAILABILITY_ZONE/INSTANCE-ID
		// ie. aws:///eu-west-1c/i-06a1d2fe9b3e8c916
		parts := strings.Split(node.Spec.ProviderID, "/")
		if len(parts) != 5 || parts[4] == "" {
			return microerror.Maskf(executionFailedError, fmt.Sprintf("invalid providerID %s in node spec %s", node.Spec.ProviderID, node.Name))
		}
		instanceID = parts[4]
	}

	{
		i := &autoscaling.TerminateInstanceInAutoScalingGroupInput{
			InstanceId:                     aws.String(instanceID),
			ShouldDecrementDesiredCapacity: aws.Bool(false),
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("terminating not healthy node %s with instanceID %s", node.Name, instanceID))

		_, err := cc.Client.TenantCluster.AWS.AutoScaling.TerminateInstanceInAutoScalingGroup(i)
		if err != nil {
			return microerror.Mask(err)
		}
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("terminated not healthy node %s with instanceID %s", node.Name, instanceID))
	}
	// expose metric about node termination
	reportNodeTermination(clusterID, node.Name, instanceID)
	return nil
}

func (r *Resource) detectBadNodes(ctx context.Context, nodes []corev1.Node) ([]corev1.Node, error) {
	var badNodes []corev1.Node
	for _, n := range nodes {
		//
		notReadyTickCount, err := r.updateNodeNotReadyTickAnnotations(ctx, n)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		if notReadyTickCount >= r.notReadyThreshold {
			badNodes = append(badNodes, n)
		}

	}

	return badNodes, nil
}

func nodeNotReady(n corev1.Node) bool {
	for _, c := range n.Status.Conditions {
		// find kubelet "ready" condition
		if c.Type == "Ready" && c.Status != "True" {
			// kubelet must be in NotReady at least for some time to avoid quick flaps
			if time.Since(c.LastHeartbeatTime.Time) >= key.NodeNotReadyDuration {
				return true
			}
		}
	}
	return false
}

// updateNodeNotReadyTickAnnotations will update annotations on the node
// depending if the node is Ready or not
// the annotation is used to track how many times node was seen as not ready
// and in case it will reach a threshold, the node will be marked for termination.
// Each reconcilation loop can increase or decrease the tick count by 1.
func (r *Resource) updateNodeNotReadyTickAnnotations(ctx context.Context, n corev1.Node) (int, error) {
	var err error
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return -1, microerror.Mask(err)
	}

	// fetch current notReady tick count from node
	// if there is no annotation yet, the value will be 0
	notReadyTickCount := 0
	{
		tick, ok := n.Annotations[key.TagNodeNotReadyTick]
		if ok {
			notReadyTickCount, err = strconv.Atoi(tick)
			if err != nil {
				return -1, microerror.Mask(err)
			}
		}
	}

	updated := false
	// increase or decrease the tick count depending on the node status
	if nodeNotReady(n) {
		notReadyTickCount++
		updated = true
	} else if notReadyTickCount > 0 {
		notReadyTickCount--
		updated = true
	}

	if updated {
		// update the tick count on the node
		n.Annotations[key.TagNodeNotReadyTick] = fmt.Sprintf("%d", notReadyTickCount)
		err = cc.Client.TenantCluster.K8s.CtrlClient().Update(ctx, &n)
		if err != nil {
			return -1, microerror.Mask(err)
		}
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated not ready tick count to %d/%d for node %s", notReadyTickCount, r.notReadyThreshold, n.Name))
	}
	return notReadyTickCount, nil
}

// maximumNodeTermination calculates the maximum number of nodes that can be terminated in single loop
// the number is calculated with help of key.NodeAutoRepairTerminationPercentage
// which determines how much percentage of nodes can be terminated
// the minimum is 1 node termination per reconciliation loop
func maximumNodeTermination(nodeCount int) int {
	limit := math.Round(float64(nodeCount) * key.NodeAutoRepairTerminationPercentage)

	if limit < 1 {
		limit = 1
	}
	return int(limit)
}
