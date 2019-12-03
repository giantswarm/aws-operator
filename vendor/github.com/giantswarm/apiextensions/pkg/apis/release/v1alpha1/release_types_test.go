package v1alpha1

import "testing"

func Test_NewReleaseCRD(t *testing.T) {
	crd := NewReleaseCRD()
	if crd == nil {
		t.Error("Release CRD was nil.")
	}
	if crd.Name == "" {
		t.Error("Release CRD name was empty.")
	}
}
