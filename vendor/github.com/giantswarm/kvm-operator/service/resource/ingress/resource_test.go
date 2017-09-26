package ingress

import (
	"context"
	"testing"

	"github.com/giantswarm/clustertpr"
	clustertprspec "github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/kvmtpr"
	"github.com/giantswarm/micrologger/microloggertest"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

func Test_Resource_Ingress_GetDesiredState(t *testing.T) {
	testCases := []struct {
		Obj               interface{}
		ExpectedAPICount  int
		ExpectedEtcdCount int
	}{
		// Test 1 ensures there is one ingress for master and worker each when there
		// is one master and one worker node in the custom object.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
						Masters: []clustertprspec.Node{
							{},
						},
						Workers: []clustertprspec.Node{
							{},
						},
					},
				},
			},
			ExpectedAPICount:  1,
			ExpectedEtcdCount: 1,
		},

		// Test 2 ensures there is one ingress for master and worker each when there
		// is one master and three worker nodes in the custom object.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
						Masters: []clustertprspec.Node{
							{},
						},
						Workers: []clustertprspec.Node{
							{},
							{},
							{},
						},
					},
				},
			},
			ExpectedAPICount:  1,
			ExpectedEtcdCount: 1,
		},

		// Test 3 ensures there is one ingress for master and worker each when there
		// are three master and three worker nodes in the custom object.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
						Masters: []clustertprspec.Node{
							{},
							{},
							{},
						},
						Workers: []clustertprspec.Node{
							{},
							{},
							{},
						},
					},
				},
			},
			ExpectedAPICount:  1,
			ExpectedEtcdCount: 1,
		},
	}

	var err error
	var newResource *Resource
	{
		resourceConfig := DefaultConfig()
		resourceConfig.K8sClient = fake.NewSimpleClientset()
		resourceConfig.Logger = microloggertest.New()
		newResource, err = New(resourceConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	for i, tc := range testCases {
		result, err := newResource.GetDesiredState(context.TODO(), tc.Obj)
		if err != nil {
			t.Fatalf("case %d expected %#v got %#v", i+1, nil, err)
		}

		ingresses, ok := result.([]*v1beta1.Ingress)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*v1beta1.Ingress{}, result)
		}

		if testGetAPICount(ingresses) != tc.ExpectedAPICount {
			t.Fatalf("case %d expected %d master nodes got %d", i+1, tc.ExpectedAPICount, testGetAPICount(ingresses))
		}

		if testGetEtcdCount(ingresses) != tc.ExpectedEtcdCount {
			t.Fatalf("case %d expected %d worker nodes got %d", i+1, tc.ExpectedEtcdCount, testGetEtcdCount(ingresses))
		}

		if len(ingresses) != tc.ExpectedAPICount+tc.ExpectedEtcdCount {
			t.Fatalf("case %d expected %d nodes got %d", i+1, tc.ExpectedAPICount+tc.ExpectedEtcdCount, len(ingresses))
		}
	}
}

func Test_Resource_Ingress_GetCreateState(t *testing.T) {
	testCases := []struct {
		Obj                  interface{}
		CurrentState         interface{}
		DesiredState         interface{}
		ExpectedIngressNames []string
	}{
		// Test 1, in case current state and desired state are empty the create
		// state should be empty.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState:         []*v1beta1.Ingress{},
			DesiredState:         []*v1beta1.Ingress{},
			ExpectedIngressNames: []string{},
		},

		// Test 2, in case current state equals desired state the create state
		// should be empty.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState: []*v1beta1.Ingress{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-1",
					},
				},
			},
			DesiredState: []*v1beta1.Ingress{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-1",
					},
				},
			},
			ExpectedIngressNames: []string{},
		},

		// Test 3, in case current state misses one item of desired state the create
		// state should contain the missing item of the desired state.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState: []*v1beta1.Ingress{},
			DesiredState: []*v1beta1.Ingress{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-1",
					},
				},
			},
			ExpectedIngressNames: []string{
				"ingress-1",
			},
		},

		// Test 4, in case current state misses items of desired state the create
		// state should contain the missing items of the desired state.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState: []*v1beta1.Ingress{},
			DesiredState: []*v1beta1.Ingress{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-2",
					},
				},
			},
			ExpectedIngressNames: []string{
				"ingress-1",
				"ingress-2",
			},
		},

		// Test 5, in case current state contains one item not being in desired
		// state the create state should not contain the missing item of the desired
		// state.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState: []*v1beta1.Ingress{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-1",
					},
				},
			},
			DesiredState:         []*v1beta1.Ingress{},
			ExpectedIngressNames: []string{},
		},

		// Test 6, in case current state contains items not being in desired state
		// the create state should not contain the missing items of the desired
		// state.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState: []*v1beta1.Ingress{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-2",
					},
				},
			},
			DesiredState:         []*v1beta1.Ingress{},
			ExpectedIngressNames: []string{},
		},

		// Test 7, in case current state contains some items of desired state the
		// create state should contain the items being in desired state which are
		// not in create state.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState: []*v1beta1.Ingress{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-2",
					},
				},
			},
			DesiredState: []*v1beta1.Ingress{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-2",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-3",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-4",
					},
				},
			},
			ExpectedIngressNames: []string{
				"ingress-3",
				"ingress-4",
			},
		},
	}

	var err error
	var newResource *Resource
	{
		resourceConfig := DefaultConfig()
		resourceConfig.K8sClient = fake.NewSimpleClientset()
		resourceConfig.Logger = microloggertest.New()
		newResource, err = New(resourceConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	for i, tc := range testCases {
		result, err := newResource.GetCreateState(context.TODO(), tc.Obj, tc.CurrentState, tc.DesiredState)
		if err != nil {
			t.Fatalf("case %d expected %#v got %#v", i+1, nil, err)
		}

		configMaps, ok := result.([]*v1beta1.Ingress)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*v1beta1.Ingress{}, result)
		}

		if len(configMaps) != len(tc.ExpectedIngressNames) {
			t.Fatalf("case %d expected %d config maps got %d", i+1, len(tc.ExpectedIngressNames), len(configMaps))
		}
	}
}

func Test_Resource_Ingress_GetDeleteState(t *testing.T) {
	testCases := []struct {
		Obj                  interface{}
		CurrentState         interface{}
		DesiredState         interface{}
		ExpectedIngressNames []string
	}{
		// Test 1, in case current state and desired state are empty the delete
		// state should be empty.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState:         []*v1beta1.Ingress{},
			DesiredState:         []*v1beta1.Ingress{},
			ExpectedIngressNames: []string{},
		},

		// Test 2, in case current state has one item and equals desired state the
		// delete state should equal the current state.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState: []*v1beta1.Ingress{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-1",
					},
				},
			},
			DesiredState: []*v1beta1.Ingress{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-1",
					},
				},
			},
			ExpectedIngressNames: []string{
				"ingress-1",
			},
		},

		// Test 3, in case current state misses one item of desired state the delete
		// state should not contain the missing item of the desired state.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState: []*v1beta1.Ingress{},
			DesiredState: []*v1beta1.Ingress{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-1",
					},
				},
			},
			ExpectedIngressNames: []string{},
		},

		// Test 4, in case current state misses items of desired state the delete
		// state should not contain the missing items of the desired state.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState: []*v1beta1.Ingress{},
			DesiredState: []*v1beta1.Ingress{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-2",
					},
				},
			},
			ExpectedIngressNames: []string{},
		},

		// Test 5, in case current state contains one item and desired state is
		// empty the delete state should be empty.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState: []*v1beta1.Ingress{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-1",
					},
				},
			},
			DesiredState:         []*v1beta1.Ingress{},
			ExpectedIngressNames: []string{},
		},

		// Test 6, in case current state contains items and desired state is empty
		// the delete state should be empty.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState: []*v1beta1.Ingress{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-2",
					},
				},
			},
			DesiredState:         []*v1beta1.Ingress{},
			ExpectedIngressNames: []string{},
		},

		// Test 7, in case all items of current state are in desired state and
		// desired state contains more items not being in current state the create
		// state should contain all items being in current state.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState: []*v1beta1.Ingress{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-2",
					},
				},
			},
			DesiredState: []*v1beta1.Ingress{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-2",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-3",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-4",
					},
				},
			},
			ExpectedIngressNames: []string{
				"ingress-1",
				"ingress-2",
			},
		},

		// Test 8, in case all items of desired state are in current state and
		// current state contains more items not being in desired state the create
		// state should contain all items being in desired state.
		{
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState: []*v1beta1.Ingress{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-2",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-3",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-4",
					},
				},
			},
			DesiredState: []*v1beta1.Ingress{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "ingress-2",
					},
				},
			},
			ExpectedIngressNames: []string{
				"ingress-1",
				"ingress-2",
			},
		},
	}

	var err error
	var newResource *Resource
	{
		resourceConfig := DefaultConfig()
		resourceConfig.K8sClient = fake.NewSimpleClientset()
		resourceConfig.Logger = microloggertest.New()
		newResource, err = New(resourceConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	for i, tc := range testCases {
		result, err := newResource.GetDeleteState(context.TODO(), tc.Obj, tc.CurrentState, tc.DesiredState)
		if err != nil {
			t.Fatalf("case %d expected %#v got %#v", i+1, nil, err)
		}

		configMaps, ok := result.([]*v1beta1.Ingress)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*v1beta1.Ingress{}, result)
		}

		if len(configMaps) != len(tc.ExpectedIngressNames) {
			t.Fatalf("case %d expected %d config maps got %d", i+1, len(tc.ExpectedIngressNames), len(configMaps))
		}
	}
}

func testGetAPICount(ingresses []*v1beta1.Ingress) int {
	var count int

	for _, i := range ingresses {
		if i.Name == "api" {
			count++
		}
	}

	return count
}

func testGetEtcdCount(ingresses []*v1beta1.Ingress) int {
	var count int

	for _, i := range ingresses {
		if i.Name == "etcd" {
			count++
		}
	}

	return count
}
