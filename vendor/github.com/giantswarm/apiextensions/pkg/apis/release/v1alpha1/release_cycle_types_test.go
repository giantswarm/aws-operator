package v1alpha1

import "testing"

func Test_NewReleaseCycleCRD(t *testing.T) {
	crd := NewReleaseCycleCRD()
	if crd == nil {
		t.Error("ReleaseCycle CRD was nil.")
	}
	if crd.Name == "" {
		t.Error("ReleaseCycle CRD name was empty.")
	}
}
