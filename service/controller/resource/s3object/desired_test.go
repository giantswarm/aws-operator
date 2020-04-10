package s3object

import (
	"context"
	"regexp"
	"strings"
	"testing"

	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/certs/certstest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/randomkeys/randomkeystest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
)

func Test_DesiredState(t *testing.T) {
	t.Parallel()
	clusterTpo := &providerv1alpha1.AWSConfig{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				label.ReleaseVersion: "1.0.0",
			},
		},
		Spec: providerv1alpha1.AWSConfigSpec{
			Cluster: providerv1alpha1.Cluster{
				ID: "test-cluster",
			},
		},
	}

	release := releasev1alpha1.NewReleaseCR()
	release.ObjectMeta.Name = "v1.0.0"
	release.Spec.Components = []releasev1alpha1.ReleaseSpecComponent{
		{
			Name:    "kubernetes",
			Version: "1.15.4",
		},
		{
			Name:    "calico",
			Version: "3.9.1",
		},
		{
			Name:    "etcd",
			Version: "3.3.15",
		},
	}
	clientset := fake.NewSimpleClientset(release)

	masterKeyPattern := "ignition/master"
	workerKeyPattern := "ignition/worker"

	masterKeyRegexp := regexp.MustCompile(masterKeyPattern)
	workerKeyRegexp := regexp.MustCompile(workerKeyPattern)

	testCases := []struct {
		obj            *providerv1alpha1.AWSConfig
		description    string
		expectedBucket string
		expectedBody   string
		template       string
	}{
		{
			description:    "case 0: basic match",
			obj:            clusterTpo,
			expectedBody:   "mybody-",
			template:       "mybody-",
			expectedBucket: "myaccountid-g8s-test-cluster",
		},
		{
			description: "case 1: template hyperkube",
			obj:         clusterTpo,
			template: `hyperkube: {{ .Images.Hyperkube }}
etcd: {{ .Images.Etcd }}
calico-cni: {{ .Images.CalicoCNI }}
calico-node: {{ .Images.CalicoNode }}
calico-kube-controllers: {{ .Images.CalicoKubeControllers }}
`,
			expectedBody: `hyperkube: example.com/giantswarm/hyperkube:v1.15.4
etcd: example.com/giantswarm/etcd:v3.3.15
calico-cni: example.com/giantswarm/cni:v3.9.1
calico-node: example.com/giantswarm/node:v3.9.1
calico-kube-controllers: example.com/giantswarm/kube-controllers:v3.9.1
`,
			expectedBucket: "myaccountid-g8s-test-cluster",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			awsClients := aws.Clients{
				KMS: &KMSClientMock{},
			}
			cloudconfig := &CloudConfigMock{
				template: tc.template,
			}

			var err error
			var newResource *Resource
			{
				c := Config{
					CertsSearcher:      certstest.NewSearcher(certstest.Config{}),
					CloudConfig:        cloudconfig,
					G8sClient:          clientset,
					Logger:             microloggertest.New(),
					RandomKeysSearcher: randomkeystest.NewSearcher(),
					RegistryDomain:     "example.com",
				}

				newResource, err = New(c)
				if err != nil {
					t.Fatal("expected", nil, "got", err)
				}
			}

			c := controllercontext.Context{
				Client: controllercontext.ContextClient{
					TenantCluster: controllercontext.ContextClientTenantCluster{
						AWS: awsClients,
					},
				},
				Status: controllercontext.ContextStatus{
					TenantCluster: controllercontext.ContextStatusTenantCluster{
						AWSAccountID: "myaccountid",
					},
				},
			}
			ctx := context.TODO()
			ctx = controllercontext.NewContext(ctx, c)

			result, err := newResource.GetDesiredState(ctx, tc.obj)
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			desiredState, ok := result.(map[string]BucketObjectState)
			if !ok {
				t.Fatalf("expected '%T', got '%T'", desiredState, result)
			}

			if len(desiredState) != 2 {
				t.Fatalf("expected 2 objects, got %d", len(desiredState))
			}

			for key, bucketObjectState := range desiredState {
				if bucketObjectState.Bucket != tc.expectedBucket {
					t.Fatalf("expected bucket %q, got %q", tc.expectedBucket, bucketObjectState.Bucket)
				}

				if bucketObjectState.Body != tc.expectedBody {
					t.Fatalf("expected body %q, got %q", tc.expectedBody, bucketObjectState.Body)
				}

				if strings.HasSuffix(key, "master") {
					if !masterKeyRegexp.MatchString(key) {
						t.Fatalf("expected key %q, to match pattern %q", key, masterKeyPattern)
					}
				} else if strings.HasSuffix(key, "worker") {
					if !workerKeyRegexp.MatchString(key) {
						t.Fatalf("expected key %q, to match pattern %q", key, workerKeyPattern)
					}
				} else {
					t.Fatalf("unexpected key %q", key)
				}
			}
		})
	}
}
