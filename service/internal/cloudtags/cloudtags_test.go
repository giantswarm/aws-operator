package cloudtags

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/v13/service/internal/unittest"
)

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
				"giantswarm.io/cluster":             "pepe2",
				"tag.provider.giantswarm.io/office": "my-office",
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

			cr := unittest.DefaultCAPIClusterWithLabels(tc.clusterID, tc.labels)
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
