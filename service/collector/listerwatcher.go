package collector

import (
	v1alpha2 "github.com/giantswarm/apiextensions/pkg/clientset/versioned/typed/infrastructure/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

// client-go uses ListerWatchers which have a slightly different method
// signature from generated interfaces so this type simply converts
// v1aplpha2.AWSClusterInterface into a cache.ListerWatcher
type ListerWatcher struct {
	clusters v1alpha2.AWSClusterInterface
}

func (lw *ListerWatcher) List(options metav1.ListOptions) (runtime.Object, error) {
	return lw.clusters.List(options)
}

func (lw *ListerWatcher) Watch(options metav1.ListOptions) (watch.Interface, error) {
	return lw.clusters.Watch(options)
}
