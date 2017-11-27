// +build k8srequired

package integration

import (
	"os"
	"testing"
)

func TestEnvVars(t *testing.T) {
	expected := "expected_value"
	actual := os.Getenv("EXPECTED_KEY")

	if expected != actual {
		t.Errorf("unexpected value for EXPECTED_KEY, expected %q, got %q", expected, actual)
	}
}
