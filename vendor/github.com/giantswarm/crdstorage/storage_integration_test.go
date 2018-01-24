// +build integration

package crdstorage

/*
	Usage:

		go test -tags=integration $(glide novendor) [FLAGS]

	Flags:

		-integration.ca string
			CA file path (default "$HOME/.minikube/ca.crt")
		-integration.crt string
			certificate file path (default "$HOME/.minikube/apiserver.crt")
		-integration.key string
			key file path (default "$HOME/.minikube/apiserver.key")
		-integration.server string
			Kubernetes API server address (default "https://$(minikube ip):8443")
*/

import (
	"context"
	"flag"
	"os/exec"
	"os/user"
	"path"
	"strings"
	"testing"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/microstorage/storagetest"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"github.com/giantswarm/operatorkit/client/k8srestconfig"
	corev1 "k8s.io/api/core/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	server  string
	crtFile string
	keyFile string
	caFile  string
)

func init() {
	u, err := user.Current()
	homePath := func(relativePath string) string {
		if err != nil {
			return ""
		}
		return path.Join(u.HomeDir, relativePath)
	}

	var serverDefault string
	{
		out, err := exec.Command("minikube", "ip").Output()
		if err == nil {
			minikubeIP := strings.TrimSpace(string(out))
			serverDefault = "https://" + string(minikubeIP) + ":8443"
		}
	}

	flag.StringVar(&server, "integration.server", serverDefault, "Kubernetes API server address")
	flag.StringVar(&crtFile, "integration.crt", homePath(".minikube/apiserver.crt"), "certificate file path")
	flag.StringVar(&keyFile, "integration.key", homePath(".minikube/apiserver.key"), "key file path")
	flag.StringVar(&caFile, "integration.ca", homePath(".minikube/ca.crt"), "CA file path")
}

func TestIntegration(t *testing.T) {
	var err error

	var restConfig *rest.Config
	{
		c := k8srestconfig.DefaultConfig()

		c.Logger = microloggertest.New()

		c.Address = server
		c.InCluster = false
		c.TLS.CAFile = caFile
		c.TLS.CrtFile = crtFile
		c.TLS.KeyFile = keyFile

		restConfig, err = k8srestconfig.New(c)
		if err != nil {
			t.Fatalf("error creating rest config: %#v", err)
		}
	}

	g8sClient, err := versioned.NewForConfig(restConfig)
	if err != nil {
		t.Fatalf("error creating g8s client: %#v", err)
	}

	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		t.Fatalf("error creating k8s client: %#v", err)
	}

	k8sExtClient, err := apiextensionsclient.NewForConfig(restConfig)
	if err != nil {
		t.Fatalf("error creating ext client: %#v", err)
	}

	var crdClient *k8scrdclient.CRDClient
	{
		c := k8scrdclient.DefaultConfig()

		c.K8sExtClient = k8sExtClient
		c.Logger = microloggertest.New()

		crdClient, err = k8scrdclient.New(c)
		if err != nil {
			t.Fatalf("error creating crd client: %#v", err)
		}
	}

	var storage *Storage
	{
		c := DefaultConfig()

		c.CRDClient = crdClient
		c.G8sClient = g8sClient
		c.K8sClient = k8sClient
		c.Logger = microloggertest.New()

		c.Name = "integration-test"
		c.Namespace = &corev1.Namespace{
			ObjectMeta: apismetav1.ObjectMeta{
				Name:      "integration-test",
				Namespace: "integration-test",
			},
		}

		storage, err = New(c)
		if err != nil {
			t.Fatalf("error creating storage: %#v", err)
		}

		defer func() {
			b := backoff.NewExponentialBackOff()
			b.MaxElapsedTime = 0
			backOff := backoff.WithMaxTries(b, 7)

			err := crdClient.EnsureDeleted(context.TODO(), v1alpha1.NewStorageConfigCRD(), backOff)
			if err != nil {
				t.Logf("error cleaning up CRD %s/%s: %#v", "integration-test", "integration-test", err)
			}
		}()
	}

	err = storage.Boot(context.TODO())
	if err != nil {
		t.Fatalf("error booting storage: %#v", err)
	}

	storagetest.Test(t, storage)
}
