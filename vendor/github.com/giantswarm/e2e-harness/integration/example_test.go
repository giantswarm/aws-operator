// +build k8srequired

package integration

import (
	"os"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestZeroInitialPods(t *testing.T) {
	cs, err := getK8sClient()
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	pods, err := cs.CoreV1().Pods("default").List(metav1.ListOptions{})
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if len(pods.Items) != 0 {
		t.Errorf("Unexpected number of pods, expected 0, got %d", len(pods.Items))
	}
}

func TestEnvVars(t *testing.T) {
	expected := "expected_value"
	actual := os.Getenv("EXPECTED_KEY")

	if expected != actual {
		t.Errorf("unexpected value for EXPECTED_KEY, expected %q, got %q", expected, actual)
	}
}
