package v1alpha1

import "testing"

func Test_NewChartCRD(t *testing.T) {
	crd := NewChartCRD()
	if crd == nil {
		t.Error("Chart CRD was nil.")
	}
}
