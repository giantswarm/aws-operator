package statusresource

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/cenkalti/backoff"
	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/errors/guest"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	// TODO
	//
	// fetch new object
	//
	// compute patches
	//
	// Apply the computed list of patches to make the status update take effect.
	// In case there are no patches we do not need to do anything here. So we
	// prevent unnecessary API calls.
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

			patches, err := r.computePatches(ctx, newAccessor, newObj)
			if err != nil {
				return microerror.Mask(err)
			}

			if len(patches) > 0 {
				err := r.applyPatches(ctx, newAccessor, patches)
				if err != nil {
					return microerror.Mask(err)
				}
			}

			return nil
		}
		b := backoff.NewExponentialBackOff()
		n := func(err error, d time.Duration) {
			r.logger.LogCtx(ctx, "level", "warning", "message", "retrying status patching due to error", "stack", fmt.Sprintf("%#v", err))
		}

		err := backoff.RetryNotify(o, b, n)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (r *Resource) applyPatches(ctx context.Context, accessor metav1.Object, patches []Patch) error {
	patches = append(patches, Patch{
		Op:    "test",
		Value: accessor.GetResourceVersion(),
		Path:  "/metadata/resourceVersion",
	})

	b, err := json.Marshal(patches)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.restClient.Patch(types.JSONPatchType).AbsPath(accessor.GetSelfLink()).Body(b).Do().Error()
	if errors.IsConflict(err) {
		return microerror.Mask(err)
	} else if errors.IsResourceExpired(err) {
		return microerror.Mask(err)
	} else if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Resource) computePatches(ctx context.Context, accessor metav1.Object, obj interface{}) ([]Patch, error) {
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
	{
		if clusterStatus.Conditions == nil && clusterStatus.Versions == nil {
			patches = append(patches, Patch{
				Op:   "add",
				Path: "/status",
				Value: Status{
					Cluster: providerv1alpha1.StatusCluster{
						Conditions: []providerv1alpha1.StatusClusterCondition{},
						Nodes:      []providerv1alpha1.StatusClusterNode{},
						Versions:   []providerv1alpha1.StatusClusterVersion{},
					},
				},
			})
		}
	}

	// Check all node versions held by the cluster status and add the version the
	// guest cluster successfully migrated to, to the historical list of versions.
	// The implication here is that an update successfully took place. This means
	// we can also add a status condition expressing the guest cluster is updated.
	{
		isNotUpdated := !clusterStatus.HasUpdatedCondition()
		sameCount := currentNodeCount != 0 && currentNodeCount == desiredNodeCount
		sameVersion := allNodesHaveVersion(clusterStatus.Nodes, desiredVersion)

		if isNotUpdated && sameCount && sameVersion {
			patches = append(patches, Patch{
				Op:    "replace",
				Path:  "/status/cluster/conditions",
				Value: clusterStatus.WithUpdatedCondition(),
			})
			patches = append(patches, Patch{
				Op:    "replace",
				Path:  "/status/cluster/versions",
				Value: clusterStatus.WithNewVersion(desiredVersion),
			})
		}
	}

	// When we notice the current and the desired guest cluster version differs,
	// an update is about to be processed. So we set the status condition
	// indicating the guest cluster is updating now.
	{
		isNotEmpty := currentVersion != ""
		isNotUpdating := !clusterStatus.HasUpdatingCondition()
		versionDiffers := currentVersion != desiredVersion

		if isNotEmpty && isNotUpdating && versionDiffers {
			patches = append(patches, Patch{
				Op:    "replace",
				Path:  "/status/cluster/conditions",
				Value: clusterStatus.WithUpdatingCondition(),
			})
		}
	}

	// Update the node status based on what the guest cluster API tells us.
	//
	// TODO this is a workaround until we can read the node status information
	// from the NodeConfig CR status. This is not possible right now because the
	// NodeConfig CRs are still used for draining by older guest clusters.
	{
		var k8sClient kubernetes.Interface
		{
			i, err := r.clusterIDFunc(obj)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			e, err := r.clusterEndpointFunc(obj)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			k8sClient, err = r.guestCluster.NewK8sClient(ctx, i, e)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}

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

				nodes = append(nodes, providerv1alpha1.StatusClusterNode{
					Name:    n,
					Version: v,
				})
			}

			nodesDiffer := !reflect.DeepEqual(clusterStatus.Nodes, nodes)

			if nodesDiffer {
				patches = append(patches, Patch{
					Op:    "replace",
					Path:  "/status/cluster/nodes",
					Value: nodes,
				})
			}
		}
	}

	// TODO emit metrics when update did not complete within a certain timeframe
	// TODO update status condition when guest cluster is migrating from creating to created status
	// TODO cancel complete reconciliation if patches were applied to prevent duplicated event processing

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
