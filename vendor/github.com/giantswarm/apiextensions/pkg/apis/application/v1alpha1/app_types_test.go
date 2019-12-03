package v1alpha1

import "testing"

func Test_NewAppCRD(t *testing.T) {
	crd := NewAppCRD()
	if crd == nil {
		t.Error("App CRD was nil.")
	}
	if crd.Name == "" {
		t.Error("App CRD name was empty.")
	}
}
