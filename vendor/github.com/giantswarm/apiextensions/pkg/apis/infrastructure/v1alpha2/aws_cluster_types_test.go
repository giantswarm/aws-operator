package v1alpha2

import "testing"

func Test_NewAWSClusterCRD(t *testing.T) {
	crd := NewAWSClusterCRD()
	if crd == nil {
		t.Error("AWSCluster CRD was nil.")
	}
	if crd.Name == "" {
		t.Error("AWSCluster CRD name was empty.")
	}
}
