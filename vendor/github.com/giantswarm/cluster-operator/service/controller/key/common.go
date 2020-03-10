package key

import (
	"fmt"

	"github.com/giantswarm/cluster-operator/pkg/label"
)

// ClusterConfigMapName returns the cluster name used in the configMap
// generated for this tenant cluster.
func ClusterConfigMapName(getter LabelsGetter) string {
	return fmt.Sprintf("%s-cluster-values", ClusterID(getter))
}

func ClusterID(getter LabelsGetter) string {
	return getter.GetLabels()[label.Cluster]
}

func IsDeleted(getter DeletionTimestampGetter) bool {
	return getter.GetDeletionTimestamp() != nil
}

func KubeConfigClusterName(getter LabelsGetter) string {
	return fmt.Sprintf("giantswarm-%s", ClusterID(getter))
}

func KubeConfigSecretName(getter LabelsGetter) string {
	return fmt.Sprintf("%s-kubeconfig", ClusterID(getter))
}

func MachineDeployment(getter LabelsGetter) string {
	return getter.GetLabels()[label.MachineDeployment]
}

func OperatorVersion(getter LabelsGetter) string {
	return getter.GetLabels()[label.OperatorVersion]
}

func OrganizationID(getter LabelsGetter) string {
	return getter.GetLabels()[label.Organization]
}

func ReleaseVersion(getter LabelsGetter) string {
	return getter.GetLabels()[label.ReleaseVersion]
}
