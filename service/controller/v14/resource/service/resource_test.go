package service

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/giantswarm/aws-operator/service/controller/v13/key"
)

func Test_toService(t *testing.T) {
	testCases := []struct {
		name          string
		input         interface{}
		expectedState *corev1.Service
		errorMatcher  func(error) bool
	}{
		{
			name: "case 0: basic match",
			input: &corev1.Service{
				ObjectMeta: apismetav1.ObjectMeta{
					Name:      "master",
					Namespace: "xy123",
					Labels: map[string]string{
						key.LabelApp:           "master",
						key.LegacyLabelCluster: "xy123",
						key.LabelCustomer:      "customer1",
						key.LabelCluster:       "xy123",
						key.LabelOrganization:  "org1",
						key.LabelVersionBundle: "1.2.3",
					},
					Annotations: map[string]string{
						key.AnnotationPrometheusCluster: "xy123",
						key.AnnotationEtcdDomain:        "etcd.cluster.mydomain:2379",
					},
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Protocol:   corev1.ProtocolTCP,
							Port:       httpsPort,
							TargetPort: intstr.FromInt(httpsPort),
						},
					},
				},
			},
			expectedState: &corev1.Service{
				ObjectMeta: apismetav1.ObjectMeta{
					Name:      "master",
					Namespace: "xy123",
					Labels: map[string]string{
						key.LabelApp:           "master",
						key.LegacyLabelCluster: "xy123",
						key.LabelCustomer:      "customer1",
						key.LabelCluster:       "xy123",
						key.LabelOrganization:  "org1",
						key.LabelVersionBundle: "1.2.3",
					},
					Annotations: map[string]string{
						key.AnnotationPrometheusCluster: "xy123",
						key.AnnotationEtcdDomain:        "etcd.cluster.mydomain:2379",
					},
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Protocol:   corev1.ProtocolTCP,
							Port:       httpsPort,
							TargetPort: intstr.FromInt(httpsPort),
						},
					},
				},
			},
		},
		{
			name: "case 1: wrong type",
			input: corev1.Service{
				ObjectMeta: apismetav1.ObjectMeta{
					Name:      "master",
					Namespace: "xy123",
					Labels: map[string]string{
						key.LabelApp:           "master",
						key.LegacyLabelCluster: "xy123",
						key.LabelCustomer:      "customer1",
						key.LabelCluster:       "xy123",
						key.LabelOrganization:  "org1",
						key.LabelVersionBundle: "1.2.3",
					},
					Annotations: map[string]string{
						key.AnnotationPrometheusCluster: "xy123",
						key.AnnotationEtcdDomain:        "etcd.cluster.mydomain:2379",
					},
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Protocol:   corev1.ProtocolTCP,
							Port:       httpsPort,
							TargetPort: intstr.FromInt(httpsPort),
						},
					},
				},
			},
			errorMatcher: IsWrongTypeError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := toService(tc.input)
			switch {
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case err != nil && !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			if !reflect.DeepEqual(result, tc.expectedState) {
				t.Fatalf("Service == %#v\n, want %#v", result, tc.expectedState)
			}
		})
	}
}

func Test_isServiceModified(t *testing.T) {
	testCases := []struct {
		name           string
		serviceA       *corev1.Service
		serviceB       *corev1.Service
		expectedResult bool
	}{
		{
			name: "case 0: basic match",
			serviceA: &corev1.Service{
				ObjectMeta: apismetav1.ObjectMeta{
					Name:      "master",
					Namespace: "xy123",
					Labels: map[string]string{
						key.LabelApp:           "master",
						key.LegacyLabelCluster: "xy123",
						key.LabelCustomer:      "customer1",
						key.LabelCluster:       "xy123",
						key.LabelOrganization:  "org1",
						key.LabelVersionBundle: "1.2.3",
					},
					Annotations: map[string]string{
						key.AnnotationPrometheusCluster: "xy123",
						key.AnnotationEtcdDomain:        "etcd.cluster.mydomain:2379",
					},
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Protocol:   corev1.ProtocolTCP,
							Port:       httpsPort,
							TargetPort: intstr.FromInt(httpsPort),
						},
					},
				},
			},
			serviceB: &corev1.Service{
				ObjectMeta: apismetav1.ObjectMeta{
					Name:      "master",
					Namespace: "xy123",
					Labels: map[string]string{
						key.LabelApp:           "master",
						key.LegacyLabelCluster: "xy123",
						key.LabelCustomer:      "customer1",
						key.LabelCluster:       "xy123",
						key.LabelOrganization:  "org1",
						key.LabelVersionBundle: "1.2.3",
					},
					Annotations: map[string]string{
						key.AnnotationPrometheusCluster: "xy123",
						key.AnnotationEtcdDomain:        "etcd.cluster.mydomain:2379",
					},
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Protocol:   corev1.ProtocolTCP,
							Port:       httpsPort,
							TargetPort: intstr.FromInt(httpsPort),
						},
					},
				},
			},
			expectedResult: false,
		},
		{
			name: "case 1: label mismatch",
			serviceA: &corev1.Service{
				ObjectMeta: apismetav1.ObjectMeta{
					Name:      "master",
					Namespace: "xy123",
					Labels: map[string]string{
						key.LabelApp:           "master",
						key.LegacyLabelCluster: "xy123",
						key.LabelCustomer:      "customer1",
						key.LabelCluster:       "xy123",
						key.LabelOrganization:  "org1",
						key.LabelVersionBundle: "1.2.3",
					},
					Annotations: map[string]string{
						key.AnnotationPrometheusCluster: "xy123",
						key.AnnotationEtcdDomain:        "etcd.cluster.mydomain:2379",
					},
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Protocol:   corev1.ProtocolTCP,
							Port:       httpsPort,
							TargetPort: intstr.FromInt(httpsPort),
						},
					},
				},
			},
			serviceB: &corev1.Service{
				ObjectMeta: apismetav1.ObjectMeta{
					Name:      "master",
					Namespace: "xy123",
					Labels: map[string]string{
						key.LabelApp:           "master",
						key.LegacyLabelCluster: "xy456",
						key.LabelCustomer:      "customer2",
						key.LabelCluster:       "xy456",
						key.LabelOrganization:  "org2",
						key.LabelVersionBundle: "1.2.4",
					},
					Annotations: map[string]string{
						key.AnnotationPrometheusCluster: "xy123",
						key.AnnotationEtcdDomain:        "etcd.cluster.mydomain:2379",
					},
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Protocol:   corev1.ProtocolTCP,
							Port:       httpsPort,
							TargetPort: intstr.FromInt(httpsPort),
						},
					},
				},
			},
			expectedResult: true,
		},
		{
			name: "case 2: annotation mismatch",
			serviceA: &corev1.Service{
				ObjectMeta: apismetav1.ObjectMeta{
					Name:      "master",
					Namespace: "xy123",
					Labels: map[string]string{
						key.LabelApp:           "master",
						key.LegacyLabelCluster: "xy123",
						key.LabelCustomer:      "customer1",
						key.LabelCluster:       "xy123",
						key.LabelOrganization:  "org1",
						key.LabelVersionBundle: "1.2.3",
					},
					Annotations: map[string]string{
						key.AnnotationPrometheusCluster: "xy123",
						key.AnnotationEtcdDomain:        "etcd.cluster.mydomain:2379",
					},
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Protocol:   corev1.ProtocolTCP,
							Port:       httpsPort,
							TargetPort: intstr.FromInt(httpsPort),
						},
					},
				},
			},
			serviceB: &corev1.Service{
				ObjectMeta: apismetav1.ObjectMeta{
					Name:      "master",
					Namespace: "xy123",
					Labels: map[string]string{
						key.LabelApp:           "master",
						key.LegacyLabelCluster: "xy123",
						key.LabelCustomer:      "customer1",
						key.LabelCluster:       "xy123",
						key.LabelOrganization:  "org1",
						key.LabelVersionBundle: "1.2.3",
					},
					Annotations: map[string]string{
						key.AnnotationPrometheusCluster: "xy456",
						key.AnnotationEtcdDomain:        "etcd2.cluster.NOTmydomain:433",
					},
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Protocol:   corev1.ProtocolTCP,
							Port:       httpsPort,
							TargetPort: intstr.FromInt(httpsPort),
						},
					},
				},
			},
			expectedResult: true,
		},
		{
			name: "case 3: ports mismatch",
			serviceA: &corev1.Service{
				ObjectMeta: apismetav1.ObjectMeta{
					Name:      "master",
					Namespace: "xy123",
					Labels: map[string]string{
						key.LabelApp:           "master",
						key.LegacyLabelCluster: "xy123",
						key.LabelCustomer:      "customer1",
						key.LabelCluster:       "xy123",
						key.LabelOrganization:  "org1",
						key.LabelVersionBundle: "1.2.3",
					},
					Annotations: map[string]string{
						key.AnnotationPrometheusCluster: "xy123",
						key.AnnotationEtcdDomain:        "etcd.cluster.mydomain:2379",
					},
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Protocol:   corev1.ProtocolTCP,
							Port:       httpsPort,
							TargetPort: intstr.FromInt(httpsPort),
						},
					},
				},
			},
			serviceB: &corev1.Service{
				ObjectMeta: apismetav1.ObjectMeta{
					Name:      "master",
					Namespace: "xy123",
					Labels: map[string]string{
						key.LabelApp:           "master",
						key.LegacyLabelCluster: "xy123",
						key.LabelCustomer:      "customer1",
						key.LabelCluster:       "xy123",
						key.LabelOrganization:  "org1",
						key.LabelVersionBundle: "1.2.3",
					},
					Annotations: map[string]string{
						key.AnnotationPrometheusCluster: "xy123",
						key.AnnotationEtcdDomain:        "etcd.cluster.mydomain:2379",
					},
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Protocol:   corev1.ProtocolTCP,
							Port:       httpsPort,
							TargetPort: intstr.FromInt(httpsPort),
						},
						{
							Protocol:   corev1.ProtocolTCP,
							Port:       89,
							TargetPort: intstr.FromInt(89),
						},
					},
				},
			},
			expectedResult: true,
		},
		{
			name: "case 4: service type mismatch",
			serviceA: &corev1.Service{
				ObjectMeta: apismetav1.ObjectMeta{
					Name:      "master",
					Namespace: "xy123",
					Labels: map[string]string{
						key.LabelApp:           "master",
						key.LegacyLabelCluster: "xy123",
						key.LabelCustomer:      "customer1",
						key.LabelCluster:       "xy123",
						key.LabelOrganization:  "org1",
						key.LabelVersionBundle: "1.2.3",
					},
					Annotations: map[string]string{
						key.AnnotationPrometheusCluster: "xy123",
						key.AnnotationEtcdDomain:        "etcd.cluster.mydomain:2379",
					},
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Protocol:   corev1.ProtocolTCP,
							Port:       httpsPort,
							TargetPort: intstr.FromInt(httpsPort),
						},
					},
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
			serviceB: &corev1.Service{
				ObjectMeta: apismetav1.ObjectMeta{
					Name:      "master",
					Namespace: "xy123",
					Labels: map[string]string{
						key.LabelApp:           "master",
						key.LegacyLabelCluster: "xy123",
						key.LabelCustomer:      "customer1",
						key.LabelCluster:       "xy123",
						key.LabelOrganization:  "org1",
						key.LabelVersionBundle: "1.2.3",
					},
					Annotations: map[string]string{
						key.AnnotationPrometheusCluster: "xy123",
						key.AnnotationEtcdDomain:        "etcd.cluster.mydomain:2379",
					},
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Protocol:   corev1.ProtocolTCP,
							Port:       httpsPort,
							TargetPort: intstr.FromInt(httpsPort),
						},
					},
					Type: corev1.ServiceTypeNodePort,
				},
			},
			expectedResult: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isServiceModified(tc.serviceA, tc.serviceB)

			if result != tc.expectedResult {
				t.Fatalf("isServiceModified '%s' failed, got %t, want %t", tc.name, result, tc.expectedResult)
			}
		})
	}
}

func Test_portsEqual(t *testing.T) {
	testCases := []struct {
		name           string
		serviceA       *corev1.Service
		serviceB       *corev1.Service
		expectedResult bool
	}{
		{
			name: "case 0: basic match",
			serviceA: &corev1.Service{
				ObjectMeta: apismetav1.ObjectMeta{
					Name:      "service1",
					Namespace: "xy123",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Protocol:   corev1.ProtocolTCP,
							Port:       443,
							TargetPort: intstr.FromInt(443),
						},
					},
				},
			},
			serviceB: &corev1.Service{
				ObjectMeta: apismetav1.ObjectMeta{
					Name:      "service1",
					Namespace: "xy123",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Protocol:   corev1.ProtocolTCP,
							Port:       443,
							TargetPort: intstr.FromInt(443),
						},
					},
				},
			},
			expectedResult: true,
		},
		{
			name: "case 1: port count mismatch",
			serviceA: &corev1.Service{
				ObjectMeta: apismetav1.ObjectMeta{
					Name:      "service1",
					Namespace: "xy123",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Protocol:   corev1.ProtocolTCP,
							Port:       443,
							TargetPort: intstr.FromInt(443),
						},
					},
				},
			},
			serviceB: &corev1.Service{
				ObjectMeta: apismetav1.ObjectMeta{
					Name:      "service1",
					Namespace: "xy123",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Protocol:   corev1.ProtocolTCP,
							Port:       443,
							TargetPort: intstr.FromInt(443),
						},
						{
							Protocol:   corev1.ProtocolTCP,
							Port:       442,
							TargetPort: intstr.FromInt(442),
						},
					},
				},
			},
			expectedResult: false,
		},
		{
			name: "case 2: port count mismatch 2",
			serviceA: &corev1.Service{
				ObjectMeta: apismetav1.ObjectMeta{
					Name:      "service1",
					Namespace: "xy123",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{},
				},
			},
			serviceB: &corev1.Service{
				ObjectMeta: apismetav1.ObjectMeta{
					Name:      "service1",
					Namespace: "xy123",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Protocol:   corev1.ProtocolTCP,
							Port:       443,
							TargetPort: intstr.FromInt(443),
						},
					},
				},
			},
			expectedResult: false,
		},
		{
			name: "case 3: protocol mismatch",
			serviceA: &corev1.Service{
				ObjectMeta: apismetav1.ObjectMeta{
					Name:      "service1",
					Namespace: "xy123",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Protocol:   corev1.ProtocolUDP,
							Port:       443,
							TargetPort: intstr.FromInt(443),
						},
					},
				},
			},
			serviceB: &corev1.Service{
				ObjectMeta: apismetav1.ObjectMeta{
					Name:      "service1",
					Namespace: "xy123",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Protocol:   corev1.ProtocolTCP,
							Port:       443,
							TargetPort: intstr.FromInt(443),
						},
					},
				},
			},
			expectedResult: false,
		},
		{
			name: "case 4: port number mismatch",
			serviceA: &corev1.Service{
				ObjectMeta: apismetav1.ObjectMeta{
					Name:      "service1",
					Namespace: "xy123",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Protocol:   corev1.ProtocolTCP,
							Port:       443,
							TargetPort: intstr.FromInt(443),
						},
					},
				},
			},
			serviceB: &corev1.Service{
				ObjectMeta: apismetav1.ObjectMeta{
					Name:      "service1",
					Namespace: "xy123",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Protocol:   corev1.ProtocolTCP,
							Port:       445,
							TargetPort: intstr.FromInt(443),
						},
					},
				},
			},
			expectedResult: false,
		},
		{
			name: "case 5: targetPort number mismatch",
			serviceA: &corev1.Service{
				ObjectMeta: apismetav1.ObjectMeta{
					Name:      "service1",
					Namespace: "xy123",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Protocol:   corev1.ProtocolTCP,
							Port:       443,
							TargetPort: intstr.FromInt(443),
						},
					},
				},
			},
			serviceB: &corev1.Service{
				ObjectMeta: apismetav1.ObjectMeta{
					Name:      "service1",
					Namespace: "xy123",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Protocol:   corev1.ProtocolTCP,
							Port:       443,
							TargetPort: intstr.FromInt(445),
						},
					},
				},
			},
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := portsEqual(tc.serviceA, tc.serviceB)

			if result != tc.expectedResult {
				t.Fatalf("portEqual '%s' failed, got %t, want %t", tc.name, result, tc.expectedResult)
			}
		})
	}
}
