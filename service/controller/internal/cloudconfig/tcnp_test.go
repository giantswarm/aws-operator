package cloudconfig

import (
	"bytes"
	"context"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"testing"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v6/pkg/template"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/randomkeys"
	"github.com/google/go-cmp/cmp"

	"github.com/giantswarm/aws-operator/service/controller/internal/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/internal/unittest"
)

// Test_Controller_CloudConfig_TCNP_Template_Render tests tenant cluster Cloud
// Config template rendering. It is meant to be used as a tool to easily check
// resulting Cloud Config template and prevent from accidental Cloud Config
// template changes.
//
// It uses a golden file as a reference template and when changes to template are
// intentional, they can be updated by providing -update flag for go test.
//
//  go test ./service/controller/internal/cloudconfig -run Test_Controller_CloudConfig_TCNP_Template_Render -update
//
func Test_Controller_CloudConfig_TCNP_Template_Render(t *testing.T) {
	testCases := []struct {
		name   string
		ctx    context.Context
		cr     infrastructurev1alpha2.AWSCluster
		certs  certs.Cluster
		images k8scloudconfig.Images
		keys   randomkeys.Cluster
		labels string
	}{
		{
			name:   "case 0: tcnp test",
			ctx:    unittest.DefaultContext(),
			cr:     unittest.DefaultCluster(),
			certs:  unittest.DefaultCerts(),
			images: unittest.DefaultImages(),
			keys:   unittest.DefaultKeys(),
			labels: "k1=v1,k2=v2",
		},
		{
			name:   "case 1: tcnp test - china",
			ctx:    unittest.ChinaContext(),
			cr:     unittest.ChinaCluster(),
			certs:  unittest.DefaultCerts(),
			images: unittest.DefaultImages(),
			keys:   unittest.DefaultKeys(),
			labels: "k1=v1,k2=v2",
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			var tcnp *TCNP
			{
				ignitionPath, err := k8scloudconfig.GetPackagePath()
				if err != nil {
					t.Fatal(err)
				}

				c := TCNPConfig{
					Config{
						Encrypter: &encrypter.EncrypterMock{},
						Logger:    microloggertest.New(),

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

				tcnp, err = NewTCNP(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			templateBody, err := tcnp.Render(tc.ctx, tc.cr, tc.certs, tc.keys, tc.images, tc.labels)
			if err != nil {
				t.Fatal(err)
			}

			p := filepath.Join("testdata", unittest.NormalizeFileName(tc.name)+".golden")

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

			if !bytes.Equal(templateBody, goldenFile) {
				t.Fatalf("\n\n%s\n", cmp.Diff(string(goldenFile), string(templateBody)))
			}
		})
	}
}
