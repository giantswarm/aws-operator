package recorder

import (
	"context"

	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclienttest"
	clientv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	pkgruntime "k8s.io/apimachinery/pkg/runtime"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
)

type Config struct {
	Component string
	K8sClient k8sclient.Interface
}

type Recorder struct {
	record.EventRecorder
}

// New creates an event recorder to send custom events to Kubernetes to be recorded for targeted Kubernetes objects.
func New(c Config) Interface {
	eventBroadcaster := record.NewBroadcaster()
	_, isfake := c.K8sClient.(*k8sclienttest.Clients)
	if !isfake {
		eventBroadcaster.StartRecordingToSink(
			&typedcorev1.EventSinkImpl{
				Interface: c.K8sClient.K8sClient().CoreV1().Events(""),
			},
		)
	}
	return &Recorder{
		eventBroadcaster.NewRecorder(c.K8sClient.Scheme(), clientv1.EventSource{Component: c.Component}),
	}
}

// Emit writes only informative events like status of creation or updates.
// Error events will be handled by operatorkit when using microerror.
func (r *Recorder) Emit(ctx context.Context, obj pkgruntime.Object, reason, message string) {
	r.Event(obj, corev1.EventTypeNormal, reason, message)
}
