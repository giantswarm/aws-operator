package awscnicleaner

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	appsv1 "k8s.io/api/apps/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	Name = "awscnicleaner"
)

type Config struct {
	CtrlClient client.Client
	Logger     micrologger.Logger
}

type objectToBeDeleted func() client.Object

// Resource that ensures the `aws-node` resources are deleted from the cluster after migration to cilium is successful
type Resource struct {
	ctrlClient client.Client
	logger     micrologger.Logger

	objectsToBeDeleted []objectToBeDeleted
}

func New(config Config) (*Resource, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	objectsToBeDeleted := []objectToBeDeleted{
		func() client.Object {
			return &v1beta1.PodSecurityPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind: "PodSecurityPolicy",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "aws-cni",
				},
			}
		},
		func() client.Object {
			return &rbacv1.ClusterRole{
				TypeMeta: metav1.TypeMeta{
					Kind: "ClusterRole",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "aws-node",
				},
			}
		},
		func() client.Object {
			return &corev1.ServiceAccount{
				TypeMeta: metav1.TypeMeta{
					Kind: "ServiceAccount",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aws-node",
					Namespace: "kube-system",
				},
			}
		},
		func() client.Object {
			return &rbacv1.ClusterRoleBinding{
				TypeMeta: metav1.TypeMeta{
					Kind: "ClusterRoleBinding",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "aws-node",
				},
			}
		},
		func() client.Object {
			return &appsv1.DaemonSet{
				TypeMeta: metav1.TypeMeta{
					Kind: "DaemonSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aws-node",
					Namespace: "kube-system",
				},
			}
		},
		func() client.Object {
			return &apiextensionsv1.CustomResourceDefinition{
				TypeMeta: metav1.TypeMeta{
					Kind: "CustomResourceDefinition",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eniconfigs.crd.k8s.amazonaws.com",
				},
			}
		},
		func() client.Object {
			return &corev1.ServiceAccount{
				TypeMeta: metav1.TypeMeta{
					Kind: "ServiceAccount",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aws-cni-restarter",
					Namespace: "kube-system",
				},
			}
		},
		func() client.Object {
			return &rbacv1.Role{
				TypeMeta: metav1.TypeMeta{
					Kind: "Role",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aws-cni-restarter",
					Namespace: "kube-system",
				},
			}
		},
		func() client.Object {
			return &rbacv1.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					Kind: "RoleBinding",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aws-cni-restarter-binding",
					Namespace: "kube-system",
				},
			}
		},
		func() client.Object {
			return &v1beta1.PodSecurityPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind: "PodSecurityPolicy",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "aws-cni-restarter",
				},
			}
		},
		func() client.Object {
			return &networkingv1.NetworkPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind: "NetworkPolicy",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aws-cni-restarter",
					Namespace: "kube-system",
				},
			}
		},
		func() client.Object {
			return &batchv1beta1.CronJob{
				TypeMeta: metav1.TypeMeta{
					Kind: "CronJob",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aws-cni-restarter",
					Namespace: "kube-system",
				},
			}
		},
	}

	r := &Resource{
		ctrlClient:         config.CtrlClient,
		logger:             config.Logger,
		objectsToBeDeleted: objectsToBeDeleted,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
