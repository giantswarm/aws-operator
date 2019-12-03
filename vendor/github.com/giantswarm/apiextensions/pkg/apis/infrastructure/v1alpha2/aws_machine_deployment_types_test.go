package v1alpha2

import "testing"

func Test_NewAWSMachineDeploymentCRD(t *testing.T) {
	crd := NewAWSMachineDeploymentCRD()
	if crd == nil {
		t.Error("AWSMachineDeployment CRD was nil.")
	}
	if crd.Name == "" {
		t.Error("AWSMachineDeployment CRD name was empty.")
	}
}
