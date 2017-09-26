package pvc

import (
	"context"
	"strings"
	"testing"

	"github.com/giantswarm/clustertpr"
	clustertprspec "github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/kvmtpr"
	kvmtprspec "github.com/giantswarm/kvmtpr/spec"
	kvmtprspeckvm "github.com/giantswarm/kvmtpr/spec/kvm"
	"github.com/giantswarm/micrologger/microloggertest"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func Test_Resource_PVC_GetDesiredState(t *testing.T) {
	testCases := []struct {
		Obj               interface{}
		ExpectedEtcdCount int
	}{
		// Test 1 ensures there is one PVC for each master when there is one master
		// and one worker node and storage type is 'persistentVolume' in the custom
		// object.
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
					KVM: kvmtprspec.KVM{
						K8sKVM: kvmtprspeckvm.K8sKVM{
							StorageType: "persistentVolume",
						},
					},
				},
			},
			ExpectedEtcdCount: 1,
		},

		// Test 2 ensures there is one PVC for each master when there is one master
		// and three worker nodes and storage type is 'persistentVolume' in the
		// custom object.
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
					KVM: kvmtprspec.KVM{
						K8sKVM: kvmtprspeckvm.K8sKVM{
							StorageType: "persistentVolume",
						},
					},
				},
			},
			ExpectedEtcdCount: 1,
		},

		// Test 3 ensures there is one PVC for each master when there are three
		// master and three worker nodes and storage type is 'persistentVolume' in
		// the custom object.
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
					KVM: kvmtprspec.KVM{
						K8sKVM: kvmtprspeckvm.K8sKVM{
							StorageType: "persistentVolume",
						},
					},
				},
			},
			ExpectedEtcdCount: 3,
		},

		// Test 4 ensures there is no PVC for each master when there is one master
		// and one worker node and storage type is 'hostPath' in the custom
		// object.
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
					KVM: kvmtprspec.KVM{
						K8sKVM: kvmtprspeckvm.K8sKVM{
							StorageType: "hostPath",
						},
					},
				},
			},
			ExpectedEtcdCount: 0,
		},

		// Test 5 ensures there is no PVC for each master when there is one master
		// and three worker nodes and storage type is 'hostPath' in the
		// custom object.
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
					KVM: kvmtprspec.KVM{
						K8sKVM: kvmtprspeckvm.K8sKVM{
							StorageType: "hostPath",
						},
					},
				},
			},
			ExpectedEtcdCount: 0,
		},

		// Test 6 ensures there is no PVC for each master when there are three
		// master and three worker nodes and storage type is 'hostPath' in
		// the custom object.
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
					KVM: kvmtprspec.KVM{
						K8sKVM: kvmtprspeckvm.K8sKVM{
							StorageType: "hostPath",
						},
					},
				},
			},
			ExpectedEtcdCount: 0,
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

		PVCs, ok := result.([]*apiv1.PersistentVolumeClaim)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*apiv1.PersistentVolumeClaim{}, result)
		}

		if testGetEtcdCount(PVCs) != tc.ExpectedEtcdCount {
			t.Fatalf("case %d expected %d worker nodes got %d", i+1, tc.ExpectedEtcdCount, testGetEtcdCount(PVCs))
		}
	}
}

func Test_Resource_PVC_GetCreateState(t *testing.T) {
	testCases := []struct {
		Obj              interface{}
		CurrentState     interface{}
		DesiredState     interface{}
		ExpectedPVCNames []string
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
			CurrentState:     []*apiv1.PersistentVolumeClaim{},
			DesiredState:     []*apiv1.PersistentVolumeClaim{},
			ExpectedPVCNames: []string{},
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
			CurrentState: []*apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
			},
			DesiredState: []*apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
			},
			ExpectedPVCNames: []string{},
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
			CurrentState: []*apiv1.PersistentVolumeClaim{},
			DesiredState: []*apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
			},
			ExpectedPVCNames: []string{
				"pvc-1",
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
			CurrentState: []*apiv1.PersistentVolumeClaim{},
			DesiredState: []*apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-2",
					},
				},
			},
			ExpectedPVCNames: []string{
				"pvc-1",
				"pvc-2",
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
			CurrentState: []*apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
			},
			DesiredState:     []*apiv1.PersistentVolumeClaim{},
			ExpectedPVCNames: []string{},
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
			CurrentState: []*apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-2",
					},
				},
			},
			DesiredState:     []*apiv1.PersistentVolumeClaim{},
			ExpectedPVCNames: []string{},
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
			CurrentState: []*apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-2",
					},
				},
			},
			DesiredState: []*apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-2",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-3",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-4",
					},
				},
			},
			ExpectedPVCNames: []string{
				"pvc-3",
				"pvc-4",
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

		configMaps, ok := result.([]*apiv1.PersistentVolumeClaim)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*apiv1.PersistentVolumeClaim{}, result)
		}

		if len(configMaps) != len(tc.ExpectedPVCNames) {
			t.Fatalf("case %d expected %d config maps got %d", i+1, len(tc.ExpectedPVCNames), len(configMaps))
		}
	}
}

func Test_Resource_PVC_GetDeleteState(t *testing.T) {
	testCases := []struct {
		Obj              interface{}
		CurrentState     interface{}
		DesiredState     interface{}
		ExpectedPVCNames []string
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
			CurrentState:     []*apiv1.PersistentVolumeClaim{},
			DesiredState:     []*apiv1.PersistentVolumeClaim{},
			ExpectedPVCNames: []string{},
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
			CurrentState: []*apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
			},
			DesiredState: []*apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
			},
			ExpectedPVCNames: []string{
				"pvc-1",
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
			CurrentState: []*apiv1.PersistentVolumeClaim{},
			DesiredState: []*apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
			},
			ExpectedPVCNames: []string{},
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
			CurrentState: []*apiv1.PersistentVolumeClaim{},
			DesiredState: []*apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-2",
					},
				},
			},
			ExpectedPVCNames: []string{},
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
			CurrentState: []*apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
			},
			DesiredState:     []*apiv1.PersistentVolumeClaim{},
			ExpectedPVCNames: []string{},
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
			CurrentState: []*apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-2",
					},
				},
			},
			DesiredState:     []*apiv1.PersistentVolumeClaim{},
			ExpectedPVCNames: []string{},
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
			CurrentState: []*apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-2",
					},
				},
			},
			DesiredState: []*apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-2",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-3",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-4",
					},
				},
			},
			ExpectedPVCNames: []string{
				"pvc-1",
				"pvc-2",
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
			CurrentState: []*apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-2",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-3",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-4",
					},
				},
			},
			DesiredState: []*apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "pvc-2",
					},
				},
			},
			ExpectedPVCNames: []string{
				"pvc-1",
				"pvc-2",
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

		configMaps, ok := result.([]*apiv1.PersistentVolumeClaim)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*apiv1.PersistentVolumeClaim{}, result)
		}

		if len(configMaps) != len(tc.ExpectedPVCNames) {
			t.Fatalf("case %d expected %d config maps got %d", i+1, len(tc.ExpectedPVCNames), len(configMaps))
		}
	}
}

func testGetEtcdCount(PVCs []*apiv1.PersistentVolumeClaim) int {
	var count int

	for _, i := range PVCs {
		if strings.HasPrefix(i.Name, "pvc-master-etcd") {
			count++
		}
	}

	return count
}
