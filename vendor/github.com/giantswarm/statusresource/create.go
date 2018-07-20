package statusresource

import (
	"context"
	"encoding/json"

	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	clusterStatus, err := r.clusterStatusFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	currentVersion := clusterStatus.LatestVersion()
	desiredVersion, err := r.versionBundleVersionFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	currentNodeCount := len(clusterStatus.Nodes)
	desiredNodeCount, err := r.nodeCountFunc(obj)
	if err != nil {
		return microerror.Mask(err)
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

	// TODO emit metrics when update did not complete within a certain timeframe
	// TODO update status condition when guest cluster is migrating from creating to created status

	// Apply the computed list of patches to make the status update take effect.
	// In case there are no patches we do not need to do anything here. So we
	// prevent unnecessary API calls.
	if len(patches) > 0 {
		err := r.patchObject(ctx, accessor, patches)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (r *Resource) patchObject(ctx context.Context, accessor metav1.Object, patches []Patch) error {
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
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
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
