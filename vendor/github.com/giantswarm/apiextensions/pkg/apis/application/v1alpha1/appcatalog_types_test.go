package v1alpha1

import "testing"

func Test_NewAppCatalogCRD(t *testing.T) {
	crd := NewAppCatalogCRD()
	if crd == nil {
		t.Error("AppCatalog CRD was nil.")
	}
}
