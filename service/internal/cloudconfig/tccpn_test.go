package cloudconfig

import (
	"bytes"
	"flag"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/ghodss/yaml"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/certs/v2/pkg/certs"
	"github.com/giantswarm/certs/v2/pkg/certstest"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v6/pkg/template"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/randomkeys"
	"github.com/google/go-cmp/cmp"

	"github.com/giantswarm/aws-operator/service/internal/cloudconfig/fixture/f001"
	"github.com/giantswarm/aws-operator/service/internal/encrypter"
	"github.com/giantswarm/aws-operator/service/internal/hamaster"
	"github.com/giantswarm/aws-operator/service/internal/images"
	"github.com/giantswarm/aws-operator/service/internal/unittest"
)

var update = flag.Bool("update", false, "update .golden CF template file")

// Test_Internal_CloudConfig_TCCPN_NewTemplates tests the k8scloudconfig
// template rendering for the TCCPN stack. It is meant to be used as a tool to
// easily check resulting Cloud Config template and prevent accidental template
// changes.
//
// It uses golden file as reference template and when changes to template are
// intentional, they can be updated by providing -update flag for go test.
//
//  go test ./service/internal/cloudconfig -run Test_Internal_CloudConfig_TCCPN_NewTemplates -update
//
func Test_Internal_CloudConfig_TCCPN_NewTemplates(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "case 0: basic tccpn",
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			ctx := unittest.DefaultContextControlPlane()
			k := unittest.NewFakeK8sClient()

			var certsSearcher certs.Interface
			{
				c := certstest.Config{
					TLS: f001.MustLoadTLS(),
				}

				certsSearcher = certstest.NewSearcher(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			var h hamaster.Interface
			{
				c := hamaster.Config{
					K8sClient: k,
				}

				h, err = hamaster.New(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			var i images.Interface
			{
				c := images.Config{
					K8sClient: k,

					RegistryDomain: "dummy",
				}

				i, err = images.New(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			var randomKeysSearcher randomkeys.Interface
			{
				c := RandomKeysConfig{
					Response: map[string]randomkeys.Cluster{
						"8y5ck": {
							APIServerEncryptionKey: randomkeys.RandomKey(mustEncryptionKeyBytes()),
						},
					},
				}

				randomKeysSearcher, err = NewRandomKeys(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			var aws infrastructurev1alpha2.AWSControlPlane
			{
				aws = unittest.DefaultAWSControlPlane()
				aws.Spec.AvailabilityZones = []string{
					"eu-central-1a",
					"eu-central-1b",
					"eu-central-1c",
				}
				err = k.CtrlClient().Create(ctx, &aws)
				if err != nil {
					t.Fatal(err)
				}

				g8s := unittest.DefaultG8sControlPlane()
				g8s.Spec.Replicas = 3
				err = k.CtrlClient().Create(ctx, &g8s)
				if err != nil {
					t.Fatal(err)
				}

				cl := unittest.DefaultCluster()
				err = k.CtrlClient().Create(ctx, &cl)
				if err != nil {
					t.Fatal(err)
				}

				re := unittest.DefaultRelease()
				err = k.CtrlClient().Create(ctx, &re)
				if err != nil {
					t.Fatal(err)
				}
			}

			var ignitionPath string
			{
				ignitionPath, err = k8scloudconfig.GetPackagePath()
				if err != nil {
					t.Fatal(err)
				}
			}

			var tccpn *TCCPN
			{
				c := TCCPNConfig{
					Config: Config{
						CertsSearcher:      certsSearcher,
						Encrypter:          &encrypter.EncrypterMock{},
						HAMaster:           h,
						Images:             i,
						K8sClient:          k,
						Logger:             microloggertest.New(),
						RandomKeysSearcher: randomKeysSearcher,

						CalicoCIDR:                18,
						CalicoMTU:                 1430,
						CalicoSubnet:              "172.18.128.0",
						ClusterDomain:             "cluster.local",
						ClusterIPRange:            "172.18.192.0/22",
						DockerDaemonCIDR:          "172.18.224.1/19",
						IgnitionPath:              ignitionPath,
						ImagePullProgressDeadline: "1m",
						NetworkSetupDockerImage:   "quay.io/giantswarm/k8s-setup-network-environment",
						RegistryDomain:            "quay.io",
						SSHUserList:               "user:ssh-rsa base64==",
						SSOPublicKey:              "user:ssh-rsa base64==",
					},
				}

				tccpn, err = NewTCCPN(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			templates, err := tccpn.NewTemplates(ctx, &aws)
			if err != nil {
				t.Fatal(err)
			}

			for i, templateBody := range templates {
				_, err = yaml.YAMLToJSONStrict([]byte(templateBody))
				if err != nil {
					t.Fatal(err)
				}

				p := filepath.Join("testdata", unittest.NormalizeFileName(tc.name+"-"+strconv.Itoa(i))+".golden")

				if *update {
					err := ioutil.WriteFile(p, []byte(templateBody), 0644)
					if err != nil {
						t.Fatal(err)
					}
				}
				goldenFile, err := ioutil.ReadFile(p)
				if err != nil {
					t.Fatal(err)
				}

				if !bytes.Equal([]byte(templateBody), goldenFile) {
					t.Fatalf("\n\n%s\n", cmp.Diff(string(goldenFile), templateBody))
				}
			}
		})
	}
}

func mustEncryptionKeyBytes() []byte {
	b, ok := f001.MustLoadRandomKey().Data[randomkeys.EncryptionKey.String()]
	if ok {
		return b
	}

	return nil
}
