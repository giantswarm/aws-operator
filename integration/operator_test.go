// +build k8srequired

package integration

import (
	"bufio"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/certificatetpr"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/pkg/api/v1"
)

const (
	defaultTimeout         = 300
	awsOperatorValuesFile  = "/tmp/aws-operator-values.yaml"
	awsOperatorChartValues = `Installation:
  V1:
    Name: gauss
    Provider:
      AWS:
        Region: ${AWS_REGION}
    Secret:
      AWSOperator:
        IDRSAPub: ${IDRSA_PUB}
        SecretYaml: |
          service:
            aws:
              accesskey:
                id: ${AWS_ACCESS_KEY_ID}
                secret: ${AWS_SECRET_ACCESS_KEY}
                token: ${AWS_SESSION_TOKEN}
              hostaccesskey:
                id: ""
                secret: ""
      Registry:
        PullSecret:
          DockerConfigJSON: "{\"auths\":{\"quay.io\":{\"auth\":\"${REGISTRY_PULL_SECRET}\"}}}"
`
	awsResourceValuesFile  = "/tmp/aws-operator-values.yaml"
	awsResourceChartValues = `commonDomain: ${COMMON_DOMAIN}
clusterName: ${CLUSTER_NAME}
clusterVersion: v_0_1_0
sshPublicKey: ${IDRSA_PUB}
versionBundleVersion: ${VERSION_BUNDLE_VERSION}
aws:
  networkCIDR: "{{.NetworkCIDR}}"
  privateSubnetCIDR: "{{.PrivateSubnetCIDR}}"
  publicSubnetCIDR: "{{.PublicSubnetCIDR}}"
  region: ${AWS_REGION}
  apiHostedZone: ${AWS_API_HOSTED_ZONE}
  ingressHostedZone: ${AWS_INGRESS_HOSTED_ZONE}
  routeTable0: ${AWS_ROUTE_TABLE_0}
  routeTable1: ${AWS_ROUTE_TABLE_1}
  vpcPeerId: ${AWS_VPC_PEER_ID}
`
	certOperatorValuesFile = "/tmp/cert-operator-values.yaml"
	// operatorChartValues values required by aws-operator-chart, the environment
	// variables will be expanded before writing the contents to a file.
	certOperatorChartValues = `commonDomain: ${COMMON_DOMAIN}
clusterName: ${CLUSTER_NAME}
Installation:
  V1:
    Auth:
      Vault:
        Address: http://vault.default.svc.cluster.local:8200
        CA:
          TTL: 1440h
    Guest:
      Kubernetes:
        API:
          EndpointBase: ${COMMON_DOMAIN}
    Secret:
      CertOperator:
        SecretYaml: |
          service:
            vault:
              config:
                token: ${VAULT_TOKEN}
      Registry:
        PullSecret:
          DockerConfigJSON: "{\"auths\":{\"quay.io\":{\"auth\":\"$REGISTRY_PULL_SECRET\"}}}"
`
)

var (
	cs        kubernetes.Interface
	awsClient aWSClient
)

// TestMain allows us to have common setup and teardown steps that are run
// once for all the tests https://golang.org/pkg/testing/#hdr-Main.
func TestMain(m *testing.M) {
	var v int
	var err error
	cs, err = newK8sClient()
	if err != nil {
		v = 1
		log.Printf("unexpected error: %v\n", err)
	}
	awsClient = newAWSClient()

	if err := setUp(cs); err != nil {
		v = 1
		log.Printf("unexpected error: %v\n", err)
	}

	if v == 0 {
		v = m.Run()
	}

	tearDown(cs)

	os.Exit(v)
}

func TestGuestClusterIsCreated(t *testing.T) {
	if err := writeAWSResourceValues(awsClient); err != nil {
		t.Errorf("unexpected error writing aws-resource-lab values file: %v", err)
	}

	if err := runCmd("helm registry install quay.io/giantswarm/aws-resource-lab-chart:stable -- -n aws-resource-lab --values " + awsOperatorValuesFile); err != nil {
		t.Errorf("unexpected error installing aws-resource-lab chart: %v", err)
	}

	operatorPodName, err := podName(cs, "giantswarm", "app=aws-operator")
	if err != nil {
		t.Errorf("unexpected error getting operator pod name: %v", err)
	}

	logEntry := "cluster '${CLUSTER_NAME}' processed"
	if os.Getenv("VERSION_BUNDLE_VERSION") == "0.2.0" {
		logEntry = "creating AWS cloudformation stack: created"
	}

	if err := waitForPodLog(cs, "giantswarm", logEntry, operatorPodName); err != nil {
		t.Errorf("unexpected error waiting for guest cluster installed: %v", err)
	}
}

func setUp(cs kubernetes.Interface) error {
	if err := createGSNamespace(cs); err != nil {
		return microerror.Mask(err)
	}

	if err := installVault(cs); err != nil {
		return microerror.Mask(err)
	}

	if err := installCertOperator(cs); err != nil {
		return microerror.Mask(err)
	}

	if err := installCertResource(cs); err != nil {
		return microerror.Mask(err)
	}

	if err := installAwsOperator(cs); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func tearDown(cs kubernetes.Interface) {
	runCmd("helm delete vault --purge")
	runCmd("helm delete cert-resource-lab --purge")
	runCmd("helm delete cert-operator --purge")
	deleteGuestCluster(cs)
	runCmd("helm delete aws-resource-lab --purge")
	runCmd("helm delete aws-operator --purge")
	cs.CoreV1().
		Namespaces().
		Delete("giantswarm", &metav1.DeleteOptions{})
	cs.ExtensionsV1beta1().
		ThirdPartyResources().
		Delete(certificatetpr.Name, &metav1.DeleteOptions{})
	cs.ExtensionsV1beta1().
		ThirdPartyResources().
		Delete(awstpr.Name, &metav1.DeleteOptions{})
}

func createGSNamespace(cs kubernetes.Interface) error {
	// check if the namespace already exists
	_, err := cs.CoreV1().
		Namespaces().
		Get("giantswarm", metav1.GetOptions{})
	if err == nil {
		return nil
	}

	namespace := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "giantswarm",
		},
	}
	_, err = cs.CoreV1().
		Namespaces().
		Create(namespace)
	if err != nil {
		return microerror.Mask(err)
	}

	return waitFor(activeNamespaceFunc(cs, "giantswarm"))
}

func installVault(cs kubernetes.Interface) error {
	if err := runCmd("helm registry install quay.io/giantswarm/vaultlab-chart:stable -- --set vaultToken=${VAULT_TOKEN} -n vault"); err != nil {
		return microerror.Mask(err)
	}

	return waitFor(runningPodFunc(cs, "default", "app=vault"))
}

func installCertOperator(cs kubernetes.Interface) error {
	certOperatorChartValuesEnv := os.ExpandEnv(certOperatorChartValues)
	if err := ioutil.WriteFile(certOperatorValuesFile, []byte(certOperatorChartValuesEnv), os.ModePerm); err != nil {
		return microerror.Mask(err)
	}
	if err := runCmd("helm registry install quay.io/giantswarm/cert-operator-chart:stable -- -n cert-operator --values " + certOperatorValuesFile); err != nil {
		return microerror.Mask(err)
	}

	return waitFor(tprFunc(cs, "certconfig"))
}

func installCertResource(cs kubernetes.Interface) error {
	err := runCmd("helm registry install quay.io/giantswarm/cert-resource-lab-chart:stable -- -n cert-resource-lab --set commonDomain=${COMMON_DOMAIN} --set clusterName=${CLUSTER_NAME}")
	if err != nil {
		return microerror.Mask(err)
	}

	secretName := fmt.Sprintf("%s-api", os.Getenv("CLUSTER_NAME"))
	log.Printf("waiting for secret %v\n", secretName)
	return waitFor(secretFunc(cs, "default", secretName))
}

func installAwsOperator(cs kubernetes.Interface) error {
	awsOperatorChartValuesEnv := os.ExpandEnv(awsOperatorChartValues)
	if err := ioutil.WriteFile(awsOperatorValuesFile, []byte(awsOperatorChartValuesEnv), os.ModePerm); err != nil {
		return microerror.Mask(err)
	}
	if err := runCmd("helm registry install quay.io/giantswarm/aws-operator-chart@1.0.0-${CIRCLE_SHA1} -- -n aws-operator --values " + awsOperatorValuesFile); err != nil {
		return microerror.Mask(err)
	}

	return waitFor(tprFunc(cs, "awsconfig"))
}

func deleteGuestCluster(cs kubernetes.Interface) error {
	if err := runCmd("kubectl delete awsconfig ${CLUSTER_NAME}"); err != nil {
		return microerror.Mask(err)
	}

	operatorPodName, err := podName(cs, "giantswarm", "app=aws-operator")
	if err != nil {
		return microerror.Mask(err)
	}

	logEntry := "cluster '${CLUSTER_NAME}' deleted"
	if os.Getenv("VERSION_BUNDLE_VERSION") == "0.2.0" {
		logEntry = "deleting AWS cloudformation stack: deleted"
	}
	return waitForPodLog(cs, "giantswarm", logEntry, operatorPodName)
}

func runCmd(cmdStr string) error {
	log.Printf("Running command %v\n", cmdStr)
	cmdEnv := os.ExpandEnv(cmdStr)
	fields := strings.Fields(cmdEnv)
	cmd := exec.Command(fields[0], fields[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	return cmd.Run()
}

func waitFor(f func() error) error {
	timeout := time.After(defaultTimeout * time.Second)
	ticker := backoff.NewTicker(backoff.NewExponentialBackOff())

	for {
		select {
		case <-timeout:
			ticker.Stop()
			return microerror.Mask(waitTimeoutError)
		case <-ticker.C:
			if err := f(); err == nil {
				return nil
			}
		}
	}
}

func runningPodFunc(cs kubernetes.Interface, namespace, labelSelector string) func() error {
	return func() error {
		pods, err := cs.CoreV1().
			Pods(namespace).
			List(metav1.ListOptions{
				LabelSelector: labelSelector,
			})
		if err != nil {
			return microerror.Mask(err)
		}
		if len(pods.Items) > 1 {
			return microerror.Mask(tooManyResultsError)
		}
		pod := pods.Items[0]
		phase := pod.Status.Phase
		if phase != v1.PodRunning {
			return microerror.Maskf(unexpectedStatusPhaseError, "current status: %s", string(phase))
		}
		return nil
	}
}

func activeNamespaceFunc(cs kubernetes.Interface, name string) func() error {
	return func() error {
		ns, err := cs.CoreV1().
			Namespaces().
			Get(name, metav1.GetOptions{})

		if err != nil {
			return microerror.Mask(err)
		}

		phase := ns.Status.Phase
		if phase != v1.NamespaceActive {
			return microerror.Maskf(unexpectedStatusPhaseError, "current status: %s", string(phase))
		}

		return nil
	}
}

func secretFunc(cs kubernetes.Interface, namespace, secretName string) func() error {
	return func() error {
		_, err := cs.CoreV1().
			Secrets(namespace).
			Get(secretName, metav1.GetOptions{})
		return microerror.Mask(err)
	}
}

func tprFunc(cs kubernetes.Interface, tprName string) func() error {
	return func() error {
		// FIXME: use proper clientset call when apiextensions are in place,
		// `cs.ExtensionsV1beta1().ThirdPartyResources().Get(tprName, metav1.GetOptions{})` finding
		// the tpr is not enough for being able to create a tpo.
		return runCmd("kubectl get " + tprName)
	}
}

func waitForPodLog(cs kubernetes.Interface, namespace, needle, podName string) error {
	needle = os.ExpandEnv(needle)

	timeout := time.After(defaultTimeout * time.Second)

	req := cs.CoreV1().
		RESTClient().
		Get().
		Namespace(namespace).
		Name(podName).
		Resource("pods").
		SubResource("log").
		Param("follow", strconv.FormatBool(true))

	readCloser, err := req.Stream()
	if err != nil {
		return microerror.Mask(err)
	}
	defer readCloser.Close()

	scanner := bufio.NewScanner(readCloser)
	var lastLine string
	for scanner.Scan() {
		select {
		case <-timeout:
			return microerror.Mask(waitTimeoutError)
		default:
		}
		lastLine = scanner.Text()
		log.Print(lastLine)
		if strings.Contains(lastLine, needle) {
			return nil
		}
	}
	if err := scanner.Err(); err != nil {
		return microerror.Mask(err)
	}

	return microerror.Mask(notFoundError)
}

func podName(cs kubernetes.Interface, namespace, labelSelector string) (string, error) {
	pods, err := cs.CoreV1().
		Pods(namespace).
		List(metav1.ListOptions{
			LabelSelector: labelSelector,
		})
	if err != nil {
		return "", microerror.Mask(err)
	}
	if len(pods.Items) > 1 {
		return "", microerror.Mask(tooManyResultsError)
	}
	pod := pods.Items[0]
	return pod.Name, nil
}

func writeAWSResourceValues(awsClient aWSClient) error {
	awsResourceChartValuesEnv := os.ExpandEnv(awsResourceChartValues)
	tmpl, err := template.New("awsResource").Parse(awsResourceChartValuesEnv)
	if err != nil {
		return microerror.Mask(err)
	}

	f, err := os.Create(awsResourceValuesFile)
	if err != nil {
		return microerror.Mask(err)
	}
	defer f.Close()

	vpc, err := newAWSVPCBlock(awsClient)
	if err != nil {
		return microerror.Mask(err)
	}

	err = tmpl.Execute(f, vpc)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
