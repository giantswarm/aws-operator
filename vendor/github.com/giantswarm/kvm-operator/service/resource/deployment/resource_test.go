package deployment

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/giantswarm/clustertpr"
	clustertprspec "github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/kvm-operator/service/messagecontext"
	"github.com/giantswarm/kvmtpr"
	kvmtprspec "github.com/giantswarm/kvmtpr/spec"
	kvmtprspeckvm "github.com/giantswarm/kvmtpr/spec/kvm"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/operatorkit/framework/updateallowedcontext"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	extensionsv1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

func Test_Resource_Deployment_GetDesiredState(t *testing.T) {
	testCases := []struct {
		Obj                 interface{}
		ExpectedMasterCount int
		ExpectedWorkerCount int
	}{
		// Test 1 ensures there is one deployment for master and worker each when there
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
					KVM: kvmtprspec.KVM{
						K8sKVM: kvmtprspeckvm.K8sKVM{
							StorageType: "hostPath",
						},
						Masters: []kvmtprspeckvm.Node{
							{},
						},
						Workers: []kvmtprspeckvm.Node{
							{},
						},
					},
				},
			},
			ExpectedMasterCount: 1,
			ExpectedWorkerCount: 1,
		},

		// Test 2 ensures there is one deployment for master and worker each when there
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
					KVM: kvmtprspec.KVM{
						K8sKVM: kvmtprspeckvm.K8sKVM{
							StorageType: "hostPath",
						},
						Masters: []kvmtprspeckvm.Node{
							{},
						},
						Workers: []kvmtprspeckvm.Node{
							{},
							{},
							{},
						},
					},
				},
			},
			ExpectedMasterCount: 1,
			ExpectedWorkerCount: 3,
		},

		// Test 3 ensures there is one deployment for master and worker each when there
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
					KVM: kvmtprspec.KVM{
						K8sKVM: kvmtprspeckvm.K8sKVM{
							StorageType: "hostPath",
						},
						Masters: []kvmtprspeckvm.Node{
							{},
							{},
							{},
						},
						Workers: []kvmtprspeckvm.Node{
							{},
							{},
							{},
						},
					},
				},
			},
			ExpectedMasterCount: 3,
			ExpectedWorkerCount: 3,
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

		deployments, ok := result.([]*v1beta1.Deployment)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*v1beta1.Deployment{}, result)
		}

		if testGetMasterCount(deployments) != tc.ExpectedMasterCount {
			t.Fatalf("case %d expected %d master nodes got %d", i+1, tc.ExpectedMasterCount, testGetMasterCount(deployments))
		}

		if testGetWorkerCount(deployments) != tc.ExpectedWorkerCount {
			t.Fatalf("case %d expected %d worker nodes got %d", i+1, tc.ExpectedWorkerCount, testGetWorkerCount(deployments))
		}

		if len(deployments) != tc.ExpectedMasterCount+tc.ExpectedWorkerCount {
			t.Fatalf("case %d expected %d nodes got %d", i+1, tc.ExpectedMasterCount+tc.ExpectedWorkerCount, len(deployments))
		}
	}
}

func Test_Resource_Deployment_GetCreateState(t *testing.T) {
	testCases := []struct {
		Obj                     interface{}
		CurrentState            interface{}
		DesiredState            interface{}
		ExpectedDeploymentNames []string
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
			CurrentState:            []*v1beta1.Deployment{},
			DesiredState:            []*v1beta1.Deployment{},
			ExpectedDeploymentNames: []string{},
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
			CurrentState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
				},
			},
			DesiredState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
				},
			},
			ExpectedDeploymentNames: []string{},
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
			CurrentState: []*v1beta1.Deployment{},
			DesiredState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
				},
			},
			ExpectedDeploymentNames: []string{
				"deployment-1",
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
			CurrentState: []*v1beta1.Deployment{},
			DesiredState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
					},
				},
			},
			ExpectedDeploymentNames: []string{
				"deployment-1",
				"deployment-2",
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
			CurrentState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
				},
			},
			DesiredState:            []*v1beta1.Deployment{},
			ExpectedDeploymentNames: []string{},
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
			CurrentState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
					},
				},
			},
			DesiredState:            []*v1beta1.Deployment{},
			ExpectedDeploymentNames: []string{},
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
			CurrentState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
					},
				},
			},
			DesiredState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-3",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-4",
					},
				},
			},
			ExpectedDeploymentNames: []string{
				"deployment-3",
				"deployment-4",
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

		deployments, ok := result.([]*v1beta1.Deployment)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*v1beta1.Deployment{}, result)
		}

		if len(deployments) != len(tc.ExpectedDeploymentNames) {
			t.Fatalf("case %d expected %d config maps got %d", i+1, len(tc.ExpectedDeploymentNames), len(deployments))
		}
	}
}

func Test_Resource_Deployment_GetDeleteState(t *testing.T) {
	testCases := []struct {
		Obj                     interface{}
		CurrentState            interface{}
		DesiredState            interface{}
		ExpectedDeploymentNames []string
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
			CurrentState:            []*v1beta1.Deployment{},
			DesiredState:            []*v1beta1.Deployment{},
			ExpectedDeploymentNames: []string{},
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
			CurrentState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
				},
			},
			DesiredState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
				},
			},
			ExpectedDeploymentNames: []string{
				"deployment-1",
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
			CurrentState: []*v1beta1.Deployment{},
			DesiredState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
				},
			},
			ExpectedDeploymentNames: []string{},
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
			CurrentState: []*v1beta1.Deployment{},
			DesiredState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
					},
				},
			},
			ExpectedDeploymentNames: []string{},
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
			CurrentState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
				},
			},
			DesiredState:            []*v1beta1.Deployment{},
			ExpectedDeploymentNames: []string{},
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
			CurrentState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
					},
				},
			},
			DesiredState:            []*v1beta1.Deployment{},
			ExpectedDeploymentNames: []string{},
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
			CurrentState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
					},
				},
			},
			DesiredState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-3",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-4",
					},
				},
			},
			ExpectedDeploymentNames: []string{
				"deployment-1",
				"deployment-2",
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
			CurrentState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-3",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-4",
					},
				},
			},
			DesiredState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
					},
				},
			},
			ExpectedDeploymentNames: []string{
				"deployment-1",
				"deployment-2",
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

		deployments, ok := result.([]*v1beta1.Deployment)
		if !ok {
			t.Fatalf("case %d expected %T got %T", i+1, []*v1beta1.Deployment{}, result)
		}

		if len(deployments) != len(tc.ExpectedDeploymentNames) {
			t.Fatalf("case %d expected %d config maps got %d", i+1, len(tc.ExpectedDeploymentNames), len(deployments))
		}
	}
}

func Test_Resource_Deployment_GetUpdateState(t *testing.T) {
	testCases := []struct {
		Ctx                         context.Context
		Obj                         interface{}
		CurrentState                interface{}
		DesiredState                interface{}
		ExpectedDeploymentsToCreate []*v1beta1.Deployment
		ExpectedDeploymentsToDelete []*v1beta1.Deployment
		ExpectedDeploymentsToUpdate []*v1beta1.Deployment
	}{
		// Test 1, in case current state and desired state are empty the create,
		// delete and update state should be empty.
		{
			Ctx: context.TODO(),
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState:                []*v1beta1.Deployment{},
			DesiredState:                []*v1beta1.Deployment{},
			ExpectedDeploymentsToCreate: nil,
			ExpectedDeploymentsToDelete: nil,
			ExpectedDeploymentsToUpdate: nil,
		},

		// Test 2, in case current state and desired state are equal the create,
		// delete and update state should be empty.
		{
			Ctx: context.TODO(),
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-1-config-map-1",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			DesiredState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-1-config-map-1",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			ExpectedDeploymentsToCreate: nil,
			ExpectedDeploymentsToDelete: nil,
			ExpectedDeploymentsToUpdate: nil,
		},

		// Test 3, in case current state misses one item of desired state the delete
		// state should not contain the missing item of the desired state.
		{
			Ctx: context.TODO(),
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState: []*v1beta1.Deployment{},
			DesiredState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-1-config-map-1",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			ExpectedDeploymentsToCreate: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-1-config-map-1",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			ExpectedDeploymentsToDelete: nil,
			ExpectedDeploymentsToUpdate: nil,
		},

		// Test 4, in case current state contains two items and desired state is
		// missing one of them the delete state should contain the the missing item
		// from the current state.
		{
			Ctx: context.TODO(),
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-1-config-map-1",
												},
											},
										},
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-2-config-map-2",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			DesiredState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-1-config-map-1",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			ExpectedDeploymentsToCreate: nil,
			ExpectedDeploymentsToDelete: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-2-config-map-2",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			ExpectedDeploymentsToUpdate: nil,
		},

		// Test 5, in case current state contains two items and desired state
		// contains the same state but one object is modified internally the update
		// state should be empty in case updates are not allowed.
		{
			Ctx: context.TODO(),
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-1-config-map-1",
												},
											},
										},
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-2-container-2-modified",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-2-config-map-2",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			DesiredState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-1-config-map-1",
												},
											},
										},
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-2-config-map-2",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			ExpectedDeploymentsToCreate: nil,
			ExpectedDeploymentsToDelete: nil,
			ExpectedDeploymentsToUpdate: nil,
		},

		// Test 6, in case current state contains two items and desired state
		// contains the same state but one object is modified internally the update
		// state should contain the the modified item from the current state.
		{
			Ctx: func() context.Context {
				ctx := context.Background()

				{
					m := messagecontext.NewMessage()
					m.ConfigMapNames = append(m.ConfigMapNames, "deployment-2-config-map-2")
					ctx = messagecontext.NewContext(ctx, m)
				}

				{
					ctx = updateallowedcontext.NewContext(ctx, make(chan struct{}))
					updateallowedcontext.SetUpdateAllowed(ctx)
				}

				return ctx
			}(),
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-1-config-map-1",
												},
											},
										},
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-2-container-2-modified",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-2-config-map-2",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			DesiredState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-1-config-map-1",
												},
											},
										},
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-2-config-map-2",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			ExpectedDeploymentsToCreate: nil,
			ExpectedDeploymentsToDelete: nil,
			ExpectedDeploymentsToUpdate: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-2-config-map-2",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},

		// Test 7, same as 6 but ensuring the right deployments are computed as
		// update state when correspondig config names have changed.
		{
			Ctx: func() context.Context {
				ctx := context.Background()

				{
					m := messagecontext.NewMessage()
					m.ConfigMapNames = append(m.ConfigMapNames, "deployment-2-config-map-2")
					ctx = messagecontext.NewContext(ctx, m)
				}

				{
					ctx = updateallowedcontext.NewContext(ctx, make(chan struct{}))
					updateallowedcontext.SetUpdateAllowed(ctx)
				}

				return ctx
			}(),
			Obj: &kvmtpr.CustomObject{
				Spec: kvmtpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: clustertprspec.Cluster{
							ID: "al9qy",
						},
					},
				},
			},
			CurrentState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-1-config-map-1",
												},
											},
										},
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-1-config-map-1",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			DesiredState: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-1",
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-1-container-1",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-1-config-map-1",
												},
											},
										},
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-2-config-map-2",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			ExpectedDeploymentsToCreate: nil,
			ExpectedDeploymentsToDelete: nil,
			ExpectedDeploymentsToUpdate: []*v1beta1.Deployment{
				{
					ObjectMeta: apismetav1.ObjectMeta{
						Name: "deployment-2",
					},
					Spec: extensionsv1.DeploymentSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name: "deployment-2-container-2",
									},
								},
								Volumes: []apiv1.Volume{
									{
										Name: "cloud-config",
										VolumeSource: apiv1.VolumeSource{
											ConfigMap: &apiv1.ConfigMapVolumeSource{
												LocalObjectReference: apiv1.LocalObjectReference{
													Name: "deployment-2-config-map-2",
												},
											},
										},
									},
								},
							},
						},
					},
				},
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

	for _, tc := range testCases {
		createState, deleteState, updateState, err := newResource.GetUpdateState(tc.Ctx, tc.Obj, tc.CurrentState, tc.DesiredState)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		t.Run("deploymentsToCreate", func(t *testing.T) {
			deploymentsToCreate, ok := createState.([]*v1beta1.Deployment)
			if !ok {
				t.Fatalf("expected %T got %T", []*v1beta1.Deployment{}, createState)
			}
			if !reflect.DeepEqual(deploymentsToCreate, tc.ExpectedDeploymentsToCreate) {
				t.Fatalf("expected %#v got %#v", tc.ExpectedDeploymentsToCreate, deploymentsToCreate)
			}
		})

		t.Run("deploymentsToDelete", func(t *testing.T) {
			deploymentsToDelete, ok := deleteState.([]*v1beta1.Deployment)
			if !ok {
				t.Fatalf("expected %T got %T", []*v1beta1.Deployment{}, deleteState)
			}
			if !reflect.DeepEqual(deploymentsToDelete, tc.ExpectedDeploymentsToDelete) {
				t.Fatalf("expected %#v got %#v", tc.ExpectedDeploymentsToDelete, deploymentsToDelete)
			}
		})

		t.Run("deploymentsToUpdate", func(t *testing.T) {
			deploymentsToUpdate, ok := updateState.([]*v1beta1.Deployment)
			if !ok {
				t.Fatalf("expected %T got %T", []*v1beta1.Deployment{}, updateState)
			}
			if !reflect.DeepEqual(deploymentsToUpdate, tc.ExpectedDeploymentsToUpdate) {
				t.Fatalf("expected %#v got %#v", tc.ExpectedDeploymentsToUpdate, deploymentsToUpdate)
			}
		})
	}
}

func testGetMasterCount(deployments []*v1beta1.Deployment) int {
	var count int

	for _, d := range deployments {
		if strings.HasPrefix(d.Name, "master-") {
			count++
		}
	}

	return count
}

func testGetWorkerCount(deployments []*v1beta1.Deployment) int {
	var count int

	for _, d := range deployments {
		if strings.HasPrefix(d.Name, "worker-") {
			count++
		}
	}

	return count
}
