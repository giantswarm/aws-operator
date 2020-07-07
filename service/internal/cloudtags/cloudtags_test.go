package cloudtags

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/aws-operator/service/internal/unittest"
)

func Test_CloudTags_AreEquals(t *testing.T) {
	testCases := []struct {
		name        string
		ctags       map[string]string
		stags       map[string]string
		expectValue bool
	}{
		{
			name:        "case 0",
			ctags:       map[string]string{},
			stags:       map[string]string{},
			expectValue: true,
		},
		{
			name: "case 1",
			ctags: map[string]string{
				"key": "value",
			},
			stags: map[string]string{
				"key": "value2",
			},
			expectValue: false,
		},
		{
			name: "case 2",
			ctags: map[string]string{
				"key": "value",
			},
			stags: map[string]string{
				"key2": "value",
			},
			expectValue: false,
		},
		{
			name: "case 3",
			ctags: map[string]string{
				"key": "value",
			},
			stags: map[string]string{
				"key": "value",
			},
			expectValue: true,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			var logger micrologger.Logger
			{
				c := micrologger.Config{}

				logger, err = micrologger.New(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			var ct *CloudTags
			{
				c := Config{
					K8sClient: unittest.FakeK8sClient(),
					Logger:    logger,
				}

				ct, err = New(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			result := ct.AreClusterTagsEquals(context.Background(), tc.ctags, tc.stags)
			if result != tc.expectValue {
				t.Fatalf("expected %t got %t", tc.expectValue, result)
			}
		})
	}
}

func Test_CloudTags_GetTagsByCluster(t *testing.T) {
	testCases := []struct {
		name         string
		clusterID    string
		ctx          context.Context
		labels       map[string]string
		expectedTags map[string]string
	}{
		{
			name:      "case 0",
			clusterID: "pepe2",
			labels: map[string]string{
				"giantswarm.io/cluster": "pepe2",
			},
			ctx:          context.Background(),
			expectedTags: map[string]string{},
		},
		{
			name:      "case 1",
			clusterID: "pepe2",
			labels: map[string]string{
				"giantswarm.io/cluster": "pepe2",
				"aws-tag/office":        "my-office",
			},
			ctx: context.Background(),
			expectedTags: map[string]string{
				"office": "my-office",
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			var logger micrologger.Logger
			{
				c := micrologger.Config{}

				logger, err = micrologger.New(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			var ct *CloudTags
			{
				c := Config{
					K8sClient: unittest.FakeK8sClient(),
					Logger:    logger,
				}

				ct, err = New(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			cr := apiv1alpha2.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Labels:    tc.labels,
					Namespace: metav1.NamespaceDefault,
					Name:      tc.clusterID,
				},
				Spec: apiv1alpha2.ClusterSpec{},
			}
			err = ct.k8sClient.CtrlClient().Create(tc.ctx, &cr)
			if err != nil {
				t.Fatal(err)
			}

			result, err := ct.GetTagsByCluster(tc.ctx, tc.clusterID)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(tc.expectedTags, result) {
				t.Fatalf("expected %q got %q", tc.expectedTags, result)
			}
		})
	}
}
