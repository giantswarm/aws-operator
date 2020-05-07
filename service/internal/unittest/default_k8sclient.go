package unittest

import (
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/k8sclient/k8scrdclient"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type fakeK8sClient struct {
	ctrlClient client.Client
}

func FakeK8sClient() k8sclient.Interface {
	var err error

	var k8sClient k8sclient.Interface
	{
		scheme := runtime.NewScheme()
		err = infrastructurev1alpha2.AddToScheme(scheme)
		if err != nil {
			panic(err)
		}
		err = releasev1alpha1.AddToScheme(scheme)
		if err != nil {
			panic(err)
		}

		k8sClient = &fakeK8sClient{
			ctrlClient: fake.NewFakeClientWithScheme(scheme),
		}
	}

	return k8sClient
}

func (f *fakeK8sClient) CRDClient() k8scrdclient.Interface {
	return nil
}

func (f *fakeK8sClient) CtrlClient() client.Client {
	return f.ctrlClient
}

func (f *fakeK8sClient) DynClient() dynamic.Interface {
	return nil
}

func (f *fakeK8sClient) ExtClient() apiextensionsclient.Interface {
	return nil
}

func (f *fakeK8sClient) G8sClient() versioned.Interface {
	return nil
}

func (f *fakeK8sClient) K8sClient() kubernetes.Interface {
	return nil
}

func (f *fakeK8sClient) RESTClient() rest.Interface {
	return nil
}

func (f *fakeK8sClient) RESTConfig() *rest.Config {
	return nil
}

func (f *fakeK8sClient) Scheme() *runtime.Scheme {
	return nil
}
