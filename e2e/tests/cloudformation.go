package tests

import (
	"log"

	"github.com/giantswarm/aws-operator/e2e/k8s"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	cs, err := k8s.Client()
	if err != nil {
		log.Println("Could not create k8s client: ", err.Error())
		return
	}
	logger, err := micrologger.New(micrologger.DefaultConfig())
	if err != nil {
		log.Println("Could not create logger: ", err.Error())
		return
	}

	ts := &TestSet{
		clientset: cs,
		logger:    logger,
	}
	Add(ts.TestCRExists)
}

func (ts *TestSet) TestCRExists() (string, error) {
	desc := "cloudformation TPR exists"

	_, err := ts.clientset.ExtensionsV1beta1().ThirdPartyResources().Get("cloudformation.giantswarm.io", metav1.GetOptions{})
	if err != nil {
		ts.logger.Log("debug", "cf tpr not found, "+err.Error())
		return desc, err
	}
	return desc, nil
}
