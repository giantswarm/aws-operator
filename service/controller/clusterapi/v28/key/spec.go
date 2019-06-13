package key

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeletionTimestampGetter interface {
	GetDeletionTimestamp() *metav1.Time
}

type LabelsGetter interface {
	GetLabels() map[string]string
}
