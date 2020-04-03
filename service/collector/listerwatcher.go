package collector

import (
	v1alpha2 "github.com/giantswarm/apiextensions/pkg/clientset/versioned/typed/infrastructure/v1alpha2"
	v1alpha1 "github.com/giantswarm/apiextensions/pkg/clientset/versioned/typed/provider/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

// client-go uses ListerWatchers which have a slightly different method
// signature from generated interfaces so this type simply converts
// v1aplpha2.AWSClusterInterface into a cache.ListerWatcher
type ClusterListerWatcher struct {
	clusters v1alpha2.AWSClusterInterface
}

func (lw *ClusterListerWatcher) List(options metav1.ListOptions) (runtime.Object, error) {
	return lw.clusters.List(options)
}

func (lw *ClusterListerWatcher) Watch(options metav1.ListOptions) (watch.Interface, error) {
	return lw.clusters.Watch(options)
}

type ConfigListerWatcher struct {
	configs v1alpha1.AWSConfigInterface
}

func (lw *ConfigListerWatcher) List(options metav1.ListOptions) (runtime.Object, error) {
	return lw.configs.List(options)
}

func (lw *ConfigListerWatcher) Watch(options metav1.ListOptions) (watch.Interface, error) {
	return lw.configs.Watch(options)
}
