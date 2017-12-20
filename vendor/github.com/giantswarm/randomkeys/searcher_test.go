package randomkeys

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_fillRandomKeyFromSecret(t *testing.T) {
	isNil := func(err error) bool { return err == nil }

	testCases := []struct {
		ClusterID        string
		RandomKey        key
		Secret           *corev1.Secret
		ExpectedCluster  Cluster
		ExpectedErrMatch func(error) bool
	}{
		// 0: ok.
		{
			ClusterID: "eggs2",
			RandomKey: key("encryption"),
			Secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"clusterID":  "eggs2",
						"clusterKey": "encryption",
					},
				},
				Data: map[string][]byte{
					"encryption": []byte("test-Encryption"),
				},
			},
			ExpectedCluster: Cluster{
				APIServerEncryptionKey: []byte("test-Encryption"),
			},
			ExpectedErrMatch: isNil,
		},
		// 1: cluster ID doesn't match.
		{
			ClusterID: "eggs5",
			RandomKey: key("encryption"),
			Secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"clusterID":  "eggs2",
						"clusterKey": "encryption",
					},
				},
				Data: map[string][]byte{
					"encryption": []byte("test-Encryption"),
				},
			},
			ExpectedCluster:  Cluster{},
			ExpectedErrMatch: IsInvalidSecret,
		},
		// 2: random key doesn't match.
		{
			ClusterID: "eggs2",
			RandomKey: key("randomly"),
			Secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"clusterID":  "eggs2",
						"clusterKey": "encryption",
					},
				},
				Data: map[string][]byte{
					"encryption": []byte("test-Encryption"),
				},
			},
			ExpectedCluster:  Cluster{},
			ExpectedErrMatch: IsInvalidSecret,
		},
		// 3: encryption field missing.
		{
			ClusterID: "eggs2",
			RandomKey: key("encryption"),
			Secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"clusterID":  "eggs2",
						"clusterKey": "encryption",
					},
				},
				Data: map[string][]byte{},
			},
			ExpectedCluster:  Cluster{},
			ExpectedErrMatch: IsInvalidSecret,
		},
	}

	for i, tc := range testCases {
		var cluster Cluster
		err := fillRandomKeyFromSecret(&cluster.APIServerEncryptionKey, tc.Secret, tc.ClusterID, tc.RandomKey)
		if !tc.ExpectedErrMatch(err) {
			t.Errorf("case %d: unexpected err = %v", i, err)
			continue
		}

		// It it was error match we don't want to check TLS.
		if tc.ExpectedErrMatch != nil {
			continue
		}

		if !reflect.DeepEqual(cluster, tc.ExpectedCluster) {
			t.Errorf("case %d: expected Cluster = %v, got %v", i, tc.ExpectedCluster, cluster)
			continue
		}
	}
}
