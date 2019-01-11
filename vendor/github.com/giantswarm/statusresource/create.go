package statusresource

import (
	"context"
	"fmt"
	"reflect"
	"time"

	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/errors/guest"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	"github.com/giantswarm/tenantcluster"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	r.logger.LogCtx(ctx, "level", "debug", "message", "patching CR status")

	// We process the status updates within its own backoff here to gurantee its
	// execution independent of any eventual retries via the retry resource. It
	// might happen that the reconciled object is not the latest version so any
	// patch would fail. In case the patch fails we retry until we succeed. The
	// steps of the backoff operation are as follows.
	//
	//     Fetch latest version of runtime object.
	//     Compute patches for runtime object.
	//     Apply computed list of patches.
	//
	// In case there are no patches we do not need to do anything. So we prevent
	// unnecessary API calls.
	var modified bool
	{
		o := func() error {
			accessor, err := meta.Accessor(obj)
			if err != nil {
				return microerror.Mask(err)
			}

			newObj, err := r.restClient.Get().AbsPath(accessor.GetSelfLink()).Do().Get()
			if err != nil {
				return microerror.Mask(err)
			}

			newAccessor, err := meta.Accessor(newObj)
			if err != nil {
				return microerror.Mask(err)
			}

			patches, err := r.computeCreateEventPatches(ctx, newObj)
			if err != nil {
				return microerror.Mask(err)
			}

			if len(patches) > 0 {
				err := r.applyPatches(ctx, newAccessor, patches)
				if err != nil {
					return microerror.Mask(err)
				}

				modified = true
			}

			return nil
		}
		b := r.backOffFactory()
		n := func(err error, d time.Duration) {
			r.logger.LogCtx(ctx, "level", "warning", "message", "retrying status patching due to error", "stack", fmt.Sprintf("%#v", err))
		}

		err := backoff.RetryNotify(o, b, n)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	if modified {
		r.logger.LogCtx(ctx, "level", "debug", "message", "patched CR status")
		reconciliationcanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not patch CR status")
	}

	return nil
}

func (r *Resource) computeCreateEventPatches(ctx context.Context, obj interface{}) ([]Patch, error) {
	clusterStatus, err := r.clusterStatusFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	currentVersion := clusterStatus.LatestVersion()
	desiredVersion, err := r.versionBundleVersionFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	currentNodeCount := len(clusterStatus.Nodes)
	desiredNodeCount, err := r.nodeCountFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var patches []Patch

	// In case a CR might not have a status at all, we cannot work with it below.
	// We have to initialize it upfront to be safe. Note that we only initialize
	// fields that are managed by the statusresource library implementation. There
	// might be other properties managed by external authorities who have to
	// manage their own initialization.
	patches = ensureDefaultPatches(clusterStatus, patches)

	// After initialization the most likely implication is the tenant cluster being
	// in a creation status. In case no other conditions are given and no nodes
	// are known and no versions are set, we set the tenant cluster status to a
	// creating condition.
	{
		notCreating := !clusterStatus.HasCreatingCondition()
		conditionsEmpty := len(clusterStatus.Conditions) == 0
		nodesEmpty := len(clusterStatus.Nodes) == 0
		versionsEmpty := len(clusterStatus.Versions) == 0

		if notCreating && conditionsEmpty && nodesEmpty && versionsEmpty {
			patches = append(patches, Patch{
				Op:    "replace",
				Path:  "/status/cluster/conditions",
				Value: clusterStatus.WithCreatingCondition(),
			})

			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", providerv1alpha1.StatusClusterTypeCreating))
		}

		if clusterStatus.HasCreatingCondition() && isHeartBeatOverdue(clusterStatus.GetCreatingCondition()) {
			patches = append(patches, Patch{
				Op:    "replace",
				Path:  "/status/cluster/conditions",
				Value: clusterStatus.UpdateHeartBeatOfCreatingCondition(),
			})

			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("updating heartbeat for %#q status condition", providerv1alpha1.StatusClusterTypeCreating))
		}
	}

	// Once the tenant cluster is created we set the according status condition so
	// the cluster status reflects the transitioning from creating to created.
	{
		isCreating := clusterStatus.HasCreatingCondition()
		notCreated := !clusterStatus.HasCreatedCondition()
		sameCount := currentNodeCount != 0 && currentNodeCount == desiredNodeCount
		sameVersion := allNodesHaveVersion(clusterStatus.Nodes, desiredVersion)

		if isCreating && notCreated && sameCount && sameVersion {
			patches = append(patches, Patch{
				Op:    "replace",
				Path:  "/status/cluster/conditions",
				Value: clusterStatus.WithCreatedCondition(),
			})

			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", providerv1alpha1.StatusClusterTypeCreated))
		}

		if clusterStatus.HasCreatedCondition() && isHeartBeatOverdue(clusterStatus.GetCreatedCondition()) {
			patches = append(patches, Patch{
				Op:    "replace",
				Path:  "/status/cluster/conditions",
				Value: clusterStatus.UpdateHeartBeatOfCreatedCondition(),
			})

			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("updating heartbeat for %#q status condition", providerv1alpha1.StatusClusterTypeCreated))
		}
	}

	// When we notice the current and the desired tenant cluster version differs,
	// an update is about to be processed. So we set the status condition
	// indicating the tenant cluster is updating now.
	{
		isCreated := clusterStatus.HasCreatedCondition()
		notUpdating := !clusterStatus.HasUpdatingCondition()
		versionDiffers := currentVersion != "" && currentVersion != desiredVersion

		if isCreated && notUpdating && versionDiffers {
			patches = append(patches, Patch{
				Op:    "replace",
				Path:  "/status/cluster/conditions",
				Value: clusterStatus.WithUpdatingCondition(),
			})

			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", providerv1alpha1.StatusClusterTypeUpdating))
		}

		if clusterStatus.HasUpdatingCondition() && isHeartBeatOverdue(clusterStatus.GetUpdatingCondition()) {
			patches = append(patches, Patch{
				Op:    "replace",
				Path:  "/status/cluster/conditions",
				Value: clusterStatus.UpdateHeartBeatOfUpdatingCondition(),
			})

			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("updating heartbeat for %#q status condition", providerv1alpha1.StatusClusterTypeUpdating))
		}
	}

	// Set the status cluster condition to updated when an update successfully
	// took place. Precondition for this is the tenant cluster is updating and all
	// nodes being known and all nodes having the same versions.
	{
		isUpdating := clusterStatus.HasUpdatingCondition()
		notUpdated := !clusterStatus.HasUpdatedCondition()
		sameCount := currentNodeCount != 0 && currentNodeCount == desiredNodeCount
		sameVersion := allNodesHaveVersion(clusterStatus.Nodes, desiredVersion)

		if isUpdating && notUpdated && sameCount && sameVersion {
			patches = append(patches, Patch{
				Op:    "replace",
				Path:  "/status/cluster/conditions",
				Value: clusterStatus.WithUpdatedCondition(),
			})

			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", providerv1alpha1.StatusClusterTypeUpdated))
		}

		if clusterStatus.HasUpdatedCondition() && isHeartBeatOverdue(clusterStatus.GetUpdatedCondition()) {
			patches = append(patches, Patch{
				Op:    "replace",
				Path:  "/status/cluster/conditions",
				Value: clusterStatus.UpdateHeartBeatOfUpdatedCondition(),
			})

			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("updating heartbeat for %#q status condition", providerv1alpha1.StatusClusterTypeUpdated))
		}
	}

	// Check all node versions held by the cluster status and add the version the
	// tenant cluster successfully migrated to, to the historical list of versions.
	{
		hasTransitioned := clusterStatus.HasCreatedCondition() || clusterStatus.HasUpdatedCondition()
		notSet := !clusterStatus.HasVersion(desiredVersion)
		sameCount := currentNodeCount != 0 && currentNodeCount == desiredNodeCount
		sameVersion := allNodesHaveVersion(clusterStatus.Nodes, desiredVersion)

		if hasTransitioned && notSet && sameCount && sameVersion {
			patches = append(patches, Patch{
				Op:    "replace",
				Path:  "/status/cluster/versions",
				Value: clusterStatus.WithNewVersion(desiredVersion),
			})

			r.logger.LogCtx(ctx, "level", "info", "message", "setting status versions")
		}
	}

	// Update the node status based on what the tenant cluster API tells us.
	//
	// TODO this is a workaround until we can read the node status information
	// from the NodeConfig CR status. This is not possible right now because the
	// NodeConfig CRs are still used for draining by older tenant clusters.
	{
		var k8sClient kubernetes.Interface
		{
			r.logger.LogCtx(ctx, "level", "debug", "message", "creating Kubernetes client for tenant cluster")

			i, err := r.clusterIDFunc(obj)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			e, err := r.clusterEndpointFunc(obj)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			k8sClient, err = r.tenantCluster.NewK8sClient(ctx, i, e)
			if tenantcluster.IsTimeout(err) {
				r.logger.LogCtx(ctx, "level", "debug", "message", "did not create Kubernetes client for tenant cluster")
				r.logger.LogCtx(ctx, "level", "debug", "message", "waiting for certificates timed out")
			} else if err != nil {
				return nil, microerror.Mask(err)
			} else {
				r.logger.LogCtx(ctx, "level", "debug", "message", "created Kubernetes client for tenant cluster")
			}
		}

		if k8sClient != nil {
			o := metav1.ListOptions{}
			list, err := k8sClient.CoreV1().Nodes().List(o)
			if guest.IsAPINotAvailable(err) {
				// fall through
			} else if err != nil {
				return nil, microerror.Mask(err)
			} else {
				var nodes []providerv1alpha1.StatusClusterNode

				for _, node := range list.Items {
					l := node.GetLabels()
					n := node.GetName()

					labelProvider := "giantswarm.io/provider"
					p, ok := l[labelProvider]
					if !ok {
						return nil, microerror.Maskf(missingLabelError, labelProvider)
					}
					labelVersion := p + "-operator.giantswarm.io/version"
					v, ok := l[labelVersion]
					if !ok {
						return nil, microerror.Maskf(missingLabelError, labelVersion)
					}

					nodes = append(nodes, providerv1alpha1.NewStatusClusterNode(n, v, l))
				}

				nodesDiffer := nodes != nil && !allNodesEqual(clusterStatus.Nodes, nodes)

				if nodesDiffer {
					patches = append(patches, Patch{
						Op:    "replace",
						Path:  "/status/cluster/nodes",
						Value: nodes,
					})

					r.logger.LogCtx(ctx, "level", "info", "message", "setting status nodes")
				}
			}
		}
	}

	// TODO emit metrics when update did not complete within a certain timeframe

	return patches, nil
}

func allNodesHaveVersion(nodes []providerv1alpha1.StatusClusterNode, version string) bool {
	if len(nodes) == 0 {
		return false
	}

	for _, n := range nodes {
		if n.Version != version {
			return false
		}
	}

	return true
}

func allNodesEqual(aNodes []providerv1alpha1.StatusClusterNode, bNodes []providerv1alpha1.StatusClusterNode) bool {
	aRemoved := removeTimesFromNodes(aNodes)
	bRemoved := removeTimesFromNodes(bNodes)

	return reflect.DeepEqual(aRemoved, bRemoved)
}

func removeTimesFromNodes(nodes []providerv1alpha1.StatusClusterNode) []providerv1alpha1.StatusClusterNode {
	var newNodes []providerv1alpha1.StatusClusterNode

	for _, n := range nodes {
		newNodes = append(newNodes, providerv1alpha1.StatusClusterNode{
			Name:    n.Name,
			Version: n.Version,
		})
	}

	return newNodes
}
