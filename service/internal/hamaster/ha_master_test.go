package hamaster

import (
	"context"
	"strconv"
	"testing"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/operatorkit/v7/pkg/controller/context/cachekeycontext"
	"github.com/google/go-cmp/cmp"

	"github.com/giantswarm/aws-operator/v13/service/internal/unittest"
)

func Test_HAMaster_Caching(t *testing.T) {
	testCases := []struct {
		name          string
		ctx           context.Context
		expectCaching bool
	}{
		{
			name:          "case 0",
			ctx:           cachekeycontext.NewContext(context.Background(), "1"),
			expectCaching: true,
		},
		{
			name:          "case 1",
			ctx:           context.Background(),
			expectCaching: false,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var ms []Mapping

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

			var aws infrastructurev1alpha3.AWSControlPlane
			{
				aws = unittest.DefaultAWSControlPlane()
			}

			var g8s infrastructurev1alpha3.G8sControlPlane
			{
				g8s = unittest.DefaultG8sControlPlane()
			}

			{
				aws.Spec.AvailabilityZones = []string{"a"}
				err = h.k8sClient.CtrlClient().Create(tc.ctx, &aws)
				if err != nil {
					t.Fatal(err)
				}

				g8s.Spec.Replicas = 1
				err = h.k8sClient.CtrlClient().Create(tc.ctx, &g8s)
				if err != nil {
					t.Fatal(err)
				}
			}

			{
				cl := unittest.DefaultCluster()
				ms, err = h.Mapping(tc.ctx, &cl)
				if err != nil {
					t.Fatal(err)
				}
			}

			if len(ms) != 1 {
				t.Fatalf("expected 1 mapping got %d", len(ms))
			}
			if ms[0].AZ != "a" {
				t.Fatalf("expected AZ `a` got %#q", ms[0].AZ)
			}
			if ms[0].ID != 0 {
				t.Fatalf("expected ID `0` got `%d`", ms[0].ID)
			}

			{
				aws.Spec.AvailabilityZones = []string{"a", "b", "c"}
				err = h.k8sClient.CtrlClient().Update(tc.ctx, &aws)
				if err != nil {
					t.Fatal(err)
				}

				g8s.Spec.Replicas = 3
				err = h.k8sClient.CtrlClient().Update(tc.ctx, &g8s)
				if err != nil {
					t.Fatal(err)
				}
			}

			{
				cl := unittest.DefaultCluster()
				ms, err = h.Mapping(tc.ctx, &cl)
				if err != nil {
					t.Fatal(err)
				}
			}

			if tc.expectCaching {
				if len(ms) != 1 {
					t.Fatalf("expected 1 mapping got %d", len(ms))
				}
			} else {
				if len(ms) != 3 {
					t.Fatalf("expected 3 mapping got %d", len(ms))
				}
			}
		})
	}
}

func Test_HAMaster_Reconcile(t *testing.T) {
	testCases := []struct {
		name             string
		azs              []string
		replicas         int
		expectedMappings []Mapping
	}{
		{
			name:     "case 0",
			azs:      []string{"a"},
			replicas: 1,
			expectedMappings: []Mapping{
				{
					AZ: "a",
					ID: 0,
				},
			},
		},
		{
			name:     "case 1",
			azs:      []string{"a"},
			replicas: 3,
			expectedMappings: []Mapping{
				{
					AZ: "a",
					ID: 1,
				},
				{
					AZ: "a",
					ID: 2,
				},
				{
					AZ: "a",
					ID: 3,
				},
			},
		},
		{
			name:     "case 2",
			azs:      []string{"a", "b"},
			replicas: 3,
			expectedMappings: []Mapping{
				{
					AZ: "a",
					ID: 1,
				},
				{
					AZ: "b",
					ID: 2,
				},
				{
					AZ: "a",
					ID: 3,
				},
			},
		},
		{
			name:     "case 3",
			azs:      []string{"a", "b", "c"},
			replicas: 3,
			expectedMappings: []Mapping{
				{
					AZ: "a",
					ID: 1,
				},
				{
					AZ: "b",
					ID: 2,
				},
				{
					AZ: "c",
					ID: 3,
				},
			},
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

			var aws infrastructurev1alpha3.AWSControlPlane
			{
				aws = unittest.DefaultAWSControlPlane()
			}

			var g8s infrastructurev1alpha3.G8sControlPlane
			{
				g8s = unittest.DefaultG8sControlPlane()
			}

			{
				aws.Spec.AvailabilityZones = tc.azs
				err = h.k8sClient.CtrlClient().Create(ctx, &aws)
				if err != nil {
					t.Fatal(err)
				}

				g8s.Spec.Replicas = tc.replicas
				err = h.k8sClient.CtrlClient().Create(ctx, &g8s)
				if err != nil {
					t.Fatal(err)
				}
			}

			cl := unittest.DefaultCluster()
			ms, err := h.Mapping(ctx, &cl)
			if err != nil {
				t.Fatal(err)
			}

			{
				diff := cmp.Diff(ms, tc.expectedMappings)
				if diff != "" {
					t.Fatalf("\n\n%s\n", diff)
				}
			}
		})
	}
}
