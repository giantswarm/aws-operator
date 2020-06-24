package nodeautorepair

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
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

	nodesToTerminate := r.detectBadNodes(nodeList.Items)

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d not healthy nodes", len(nodesToTerminate)))

	for _, n := range nodesToTerminate {
		err := r.terminateNode(ctx, n)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (r *Resource) terminateNode(ctx context.Context, node corev1.Node) error {
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
		instanceID = parts[3]
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
	return nil
}

func (r *Resource) detectBadNodes(nodes []corev1.Node) []corev1.Node {
	var badNodes []corev1.Node
	for _, n := range nodes {
		if isNotReadyFor(n, r.notReadyThreshold) {
			badNodes = append(badNodes, n)
		}
	}

	return badNodes
}

func isNotReadyFor(n corev1.Node, duration time.Duration) bool {
	for _, c := range n.Status.Conditions {
		// find kubelet "ready" condition
		if c.Type == "Ready" && c.Status != "True" {
			// check for how long kubelet is not ready, if it reached duration threshold
			if time.Since(c.LastHeartbeatTime.Time) >= duration {
				return true
			}
		}
	}
	return false
}
