package framework

import (
	"fmt"
	"log"
	"os"

	"github.com/giantswarm/microerror"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/giantswarm/e2e-harness/pkg/harness"
)

const (
	// minimumNodesReady represents the minimun number of ready nodes in a guest
	// cluster to be considered healthy.
	minimumNodesReady = 3
)

type Guest struct {
	k8sClient kubernetes.Interface
}

func NewGuest() (*Guest, error) {
	var guestK8sClient kubernetes.Interface
	{
		c, err := clientcmd.BuildConfigFromFlags("", harness.DefaultKubeConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		hostK8sClient, err := kubernetes.NewForConfig(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		guestK8sClient, err = newGuestK8sClient(hostK8sClient)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	g := &Guest{
		k8sClient: guestK8sClient,
	}

	return g, nil
}

func (g *Guest) WaitForAPIDown() error {
	apiDown := func() error {
		_, err := g.k8sClient.
			CoreV1().
			Services("default").
			Get("kubernetes", metav1.GetOptions{})

		if err == nil {
			return microerror.Mask(fmt.Errorf("API up"))
		}
		log.Printf("k8s API down: %v\n", err)
		return nil
	}

	log.Printf("waiting for k8s API down\n")
	err := waitConstantFor(apiDown)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (g *Guest) WaitForGuestReady() error {
	var err error

	err = g.waitForAPIUp()
	if err != nil {
		return microerror.Mask(err)
	}

	err = g.WaitForNodesUp(minimumNodesReady)
	if err != nil {
		return microerror.Mask(err)
	}

	log.Println("Guest cluster ready")

	return nil
}

func (g *Guest) WaitForNodesUp(numberOfNodes int) error {
	nodesUp := func() error {
		res, err := g.k8sClient.
			CoreV1().
			Nodes().
			List(metav1.ListOptions{})

		if err != nil {
			log.Printf("waiting for nodes ready: %v\n", err)
			return microerror.Mask(err)
		}
		if len(res.Items) != numberOfNodes {
			log.Printf("worker nodes not found")
			return microerror.Mask(notFoundError)
		}

		for _, n := range res.Items {
			for _, c := range n.Status.Conditions {
				if c.Type == v1.NodeReady && c.Status != v1.ConditionTrue {
					log.Printf("not all worker nodes ready")
					return microerror.Mask(notFoundError)
				}
			}
		}
		return nil
	}

	return waitFor(nodesUp)
}

func (g *Guest) waitForAPIUp() error {
	apiUp := func() error {
		_, err := g.k8sClient.CoreV1().Services("default").Get("kubernetes", metav1.GetOptions{})
		if err != nil {
			log.Printf("waiting for k8s API up: %v\n", err)
			return microerror.Mask(err)
		}

		log.Println("k8s API up")
		return nil
	}

	return waitFor(apiUp)
}

func newGuestK8sClient(hostK8sClient kubernetes.Interface) (kubernetes.Interface, error) {
	n := os.ExpandEnv("${CLUSTER_NAME}-api")
	s, err := hostK8sClient.CoreV1().Secrets("default").Get(n, metav1.GetOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	c := &rest.Config{
		Host: os.ExpandEnv("https://api.${CLUSTER_NAME}.${COMMON_DOMAIN_GUEST}"),
		TLSClientConfig: rest.TLSClientConfig{
			CAData:   s.Data["ca"],
			CertData: s.Data["crt"],
			KeyData:  s.Data["key"],
		},
	}

	guestK8sClient, err := kubernetes.NewForConfig(c)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return guestK8sClient, nil
}
