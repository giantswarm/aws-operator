package cmd_test

import (
	"testing"

	"github.com/giantswarm/e2e-harness/cmd"
)

func TestSetupFlags(t *testing.T) {
	remoteFlag := cmd.SetupCmd.Flags().Lookup("remote")

	t.Run("flag exists", func(t *testing.T) {
		if remoteFlag == nil {
			t.Errorf("expected remoteFlag not nil")
		}
	})
	t.Run("flag value", func(t *testing.T) {
		actual := remoteFlag.Value.String()
		expected := "true"
		if actual != expected {
			t.Errorf("expected %s, got %s", expected, actual)
		}
	})
}
