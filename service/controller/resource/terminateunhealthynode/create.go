package terminateunhealthynode

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/giantswarm/apiextensions/v3/pkg/annotation"
	"github.com/giantswarm/badnodedetector/pkg/detector"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

const (
	nodeTerminationTickThreshold = 6
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	var err error
	cr, err := key.ToCluster(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	// check for annotation enabling the node auto repair feature
	if _, ok := cr.Annotations[annotation.NodeTerminateUnhealthy]; !ok {
		r.logger.Debugf(ctx, "terminate unhealthy node feature is not enabled for this cluster, cancelling")
		return nil
	}

	if cc.Client.TenantCluster.K8s == nil {
		r.logger.Debugf(ctx, "kubernetes clients are not available in controller context yet")
		r.logger.Debugf(ctx, "canceling resource")

		return nil
	}

	var detectorService *detector.Detector
	{
		detectorConfig := detector.Config{
			K8sClient: cc.Client.TenantCluster.K8s.CtrlClient(),
			Logger:    r.logger,

			NotReadyTickThreshold: nodeTerminationTickThreshold,
		}

		detectorService, err = detector.NewDetector(detectorConfig)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	nodesToTerminate, err := detectorService.DetectBadNodes(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(nodesToTerminate) > 0 {
		for _, n := range nodesToTerminate {
			err := r.terminateNode(ctx, n, key.ClusterID(&cr))
			if err != nil {
				return microerror.Mask(err)
			}
		}

		// reset tick counters on all nodes in cluster to have a graceful period after terminating nodes
		err := detectorService.ResetTickCounters(ctx)
		if err != nil {
			return microerror.Mask(err)
		}
		r.logger.Debugf(ctx, "resetting tick node counters on all nodes in tenant cluster")
	}

	return nil
}

func (r *Resource) terminateNode(ctx context.Context, node corev1.Node, clusterID string) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	instanceID, err := getInstanceId(node)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		i := &autoscaling.TerminateInstanceInAutoScalingGroupInput{
			InstanceId:                     aws.String(instanceID),
			ShouldDecrementDesiredCapacity: aws.Bool(false),
		}

		r.logger.Debugf(ctx, "terminating not healthy node %s with instanceID %s", node.Name, instanceID)

		_, err := cc.Client.TenantCluster.AWS.AutoScaling.TerminateInstanceInAutoScalingGroup(i)
		if err != nil {
			return microerror.Mask(err)
		}
		r.logger.Debugf(ctx, "terminated not healthy node %s with instanceID %s", node.Name, instanceID)
	}
	// expose metric about node termination
	reportNodeTermination(clusterID, node.Name, instanceID)
	return nil
}

func getInstanceId(n corev1.Node) (string, error) {
	// node.spec.providerID for AWS is in format aws:///AVAILABILITY_ZONE/INSTANCE-ID
	// ie. aws:///eu-west-1c/i-06a1d2fe9b3e8c916
	parts := strings.Split(n.Spec.ProviderID, "/")
	if len(parts) != 5 || parts[4] == "" {
		return "", microerror.Maskf(invalidProviderIDError, fmt.Sprintf("invalid providerID %s in node spec %s", n.Spec.ProviderID, n.Name))
	}
	return parts[4], nil
}
