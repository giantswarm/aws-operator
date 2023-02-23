package tccpnlifecycle

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/v14/service/controller/controllercontext"
)

const (
	controlPlaneLabel = "node-role.kubernetes.io/control-plane"
	asgTagName        = "aws:autoscaling:groupName"
	lifecycleHookName = "ControlPlaneLaunching"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	ctrlClient := cc.Client.TenantCluster.K8s.CtrlClient()

	masters := &v1.NodeList{}
	err = ctrlClient.List(ctx, masters, client.MatchingLabels{
		controlPlaneLabel: "",
	})
	if err != nil {
		return microerror.Mask(err)
	}

	for _, node := range masters.Items {
		if !util.IsNodeReady(&node) {
			r.logger.Debugf(ctx, "Node %s is not ready, not completing lifecycle hooks")
			continue
		}

		instanceId, err := getInstanceId(node)
		if err != nil {
			return microerror.Mask(err)
		}

		out, err := cc.Client.TenantCluster.AWS.EC2.DescribeTags(&ec2.DescribeTagsInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("resource-id"),
					Values: []*string{
						aws.String(instanceId),
					},
				},
				{
					Name: aws.String("key"),
					Values: []*string{
						aws.String(asgTagName),
					},
				},
			},
		})
		if err != nil {
			return microerror.Mask(err)
		}

		if len(out.Tags) != 1 {
			r.logger.Debugf(ctx, "Expected exactly one tag named %q for instance %s, got %d", asgTagName, instanceId, len(out.Tags))
			continue
		}

		r.logger.Debugf(ctx, "Completing lifecycle action for node %s (instance %s)", node.Name, instanceId)

		asgName := out.Tags[0].Value

		i := &autoscaling.CompleteLifecycleActionInput{
			AutoScalingGroupName:  asgName,
			InstanceId:            aws.String(instanceId),
			LifecycleActionResult: aws.String("CONTINUE"),
			LifecycleHookName:     aws.String(lifecycleHookName),
		}

		cc, err := controllercontext.FromContext(ctx)
		if err != nil {
			return microerror.Mask(err)
		}

		_, err = cc.Client.TenantCluster.AWS.AutoScaling.CompleteLifecycleAction(i)
		if IsNoActiveLifeCycleAction(err) {
			r.logger.Debugf(ctx, "did not find life cycle hook action for tenant cluster node %#q", instanceId)
		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func getInstanceId(n v1.Node) (string, error) {
	// node.spec.providerID for AWS is in format aws:///AVAILABILITY_ZONE/INSTANCE-ID
	// ie. aws:///eu-west-1c/i-06a1d2fe9b3e8c916
	parts := strings.Split(n.Spec.ProviderID, "/")
	if len(parts) != 5 || parts[4] == "" {
		return "", microerror.Maskf(invalidProviderIDError, fmt.Sprintf("invalid providerID %s in node spec %s", n.Spec.ProviderID, n.Name))
	}
	return parts[4], nil
}
