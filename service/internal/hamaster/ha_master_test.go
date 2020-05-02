package hamaster

import (
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/giantswarm/aws-operator/service/controller/internal/unittest"
)

func Test_HAMaster_Reconcile(t *testing.T) {
	testCases := []struct {
		name        string
		azs         []string
		replicas    int
		expectedAZs []string
		expectedIDs []int
	}{
		{
			name:        "case 0",
			azs:         []string{"a"},
			replicas:    1,
			expectedAZs: []string{"a"},
			expectedIDs: []int{0},
		},
		{
			name:        "case 1",
			azs:         []string{"a"},
			replicas:    3,
			expectedAZs: []string{"a", "a", "a"},
			expectedIDs: []int{1, 2, 3},
		},
		{
			name:        "case 2",
			azs:         []string{"a", "b"},
			replicas:    3,
			expectedAZs: []string{"a", "b", "a"},
			expectedIDs: []int{1, 2, 3},
		},
		{
			name:        "case 3",
			azs:         []string{"a", "b", "c"},
			replicas:    3,
			expectedAZs: []string{"a", "b", "c"},
			expectedIDs: []int{1, 2, 3},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			ctx := unittest.DefaultContext()

			var h *HAMaster
			{
				c := Config{
					K8sClient: unittest.FakeK8sClient(),
				}

				h, err = New(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			{
				aws := unittest.DefaultAWSControlPlane()
				aws.Spec.AvailabilityZones = tc.azs
				err = h.k8sClient.CtrlClient().Create(ctx, &aws)
				if err != nil {
					t.Fatal(err)
				}

				g8s := unittest.DefaultG8sControlPlane()
				g8s.Spec.Replicas = tc.replicas
				err = h.k8sClient.CtrlClient().Create(ctx, &g8s)
				if err != nil {
					t.Fatal(err)
				}

				cl := unittest.DefaultCluster()
				err = h.Init(ctx, &cl)
				if err != nil {
					t.Fatal(err)
				}
			}

			var azs []string
			var ids []int
			for !h.Reconciled() {
				azs = append(azs, h.AZ())
				ids = append(ids, h.ID())
				h.Next()
			}

			{
				diff := cmp.Diff(tc.expectedAZs, azs)
				if diff != "" {
					t.Fatalf("\n\n%s\n", diff)
				}
			}

			{
				diff := cmp.Diff(tc.expectedIDs, ids)
				if diff != "" {
					t.Fatalf("\n\n%s\n", diff)
				}
			}
		})
	}
}
