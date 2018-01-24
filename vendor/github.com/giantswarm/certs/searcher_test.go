package certs

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_fillTLSFromSecret(t *testing.T) {
	isNil := func(err error) bool { return err == nil }

	testCases := []struct {
		ClusterID        string
		Cert             cert
		Secret           *corev1.Secret
		ExpectedTLS      TLS
		ExpectedErrMatch func(error) bool
	}{
		// 0: ok.
		{
			ClusterID: "eggs2",
			Cert:      cert("etcd"),
			Secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"clusterID":        "eggs2",
						"clusterComponent": "etcd",
					},
				},
				Data: map[string][]byte{
					"ca":  []byte("test-CA"),
					"crt": []byte("test-crt"),
					"key": []byte("test-key"),
				},
			},
			ExpectedTLS: TLS{
				CA:  []byte("test-CA"),
				Crt: []byte("test-crt"),
				Key: []byte("test-key"),
			},
			ExpectedErrMatch: isNil,
		},
		// 1: cluster ID doesn't match.
		{
			ClusterID: "eggs5",
			Cert:      cert("etcd"),
			Secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"clusterID":        "eggs2",
						"clusterComponent": "etcd",
					},
				},
				Data: map[string][]byte{
					"ca":  []byte("test-CA"),
					"crt": []byte("test-crt"),
					"key": []byte("test-key"),
				},
			},
			ExpectedTLS:      TLS{},
			ExpectedErrMatch: IsInvalidSecret,
		},
		// 2: cert doesn't match.
		{
			ClusterID: "eggs2",
			Cert:      cert("calico"),
			Secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"clusterID":        "eggs2",
						"clusterComponent": "etcd",
					},
				},
				Data: map[string][]byte{
					"ca":  []byte("test-CA"),
					"crt": []byte("test-crt"),
					"key": []byte("test-key"),
				},
			},
			ExpectedTLS:      TLS{},
			ExpectedErrMatch: IsInvalidSecret,
		},
		// 3: ca field missing.
		{
			ClusterID: "eggs2",
			Cert:      cert("etcd"),
			Secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"clusterID":        "eggs2",
						"clusterComponent": "etcd",
					},
				},
				Data: map[string][]byte{
					"crt": []byte("test-crt"),
					"key": []byte("test-key"),
				},
			},
			ExpectedTLS:      TLS{},
			ExpectedErrMatch: IsInvalidSecret,
		},
		// 4: crt field missing.
		{
			ClusterID: "eggs2",
			Cert:      cert("etcd"),
			Secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"clusterID":        "eggs2",
						"clusterComponent": "etcd",
					},
				},
				Data: map[string][]byte{
					"ca":  []byte("test-CA"),
					"key": []byte("test-key"),
				},
			},
			ExpectedTLS:      TLS{},
			ExpectedErrMatch: IsInvalidSecret,
		},
		// 5: key field missing.
		{
			ClusterID: "eggs2",
			Cert:      cert("etcd"),
			Secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"clusterID":        "eggs2",
						"clusterComponent": "etcd",
					},
				},
				Data: map[string][]byte{
					"ca":  []byte("test-CA"),
					"crt": []byte("test-crt"),
				},
			},
			ExpectedTLS:      TLS{},
			ExpectedErrMatch: IsInvalidSecret,
		},
	}

	for i, tc := range testCases {
		var tls TLS
		err := fillTLSFromSecret(&tls, tc.Secret, tc.ClusterID, tc.Cert)
		if !tc.ExpectedErrMatch(err) {
			t.Errorf("case %d: unexpected err = %v", i, err)
			continue
		}

		// It it was error match we don't want to check TLS.
		if tc.ExpectedErrMatch != nil {
			continue
		}

		if !reflect.DeepEqual(tls, tc.ExpectedTLS) {
			t.Errorf("case %d: expected TLS = %v, got %v", i, tc.ExpectedTLS, tls)
			continue
		}
	}
}
