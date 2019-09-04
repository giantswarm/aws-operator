package cloudconfig

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/giantswarm/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v_4_7_0"
	"github.com/giantswarm/randomkeys"
	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v30/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v30/unittest"
	"github.com/giantswarm/micrologger/microloggertest"
)

var update = flag.Bool("update", false, "update .golden CF template file")

// Test_Controller_CloudConfig_TCCP_Template_Render tests tenant cluster Cloud
// Config template rendering. It is meant to be used as a tool to easily check
// resulting Cloud Config template and prevent from accidental Cloud Config
// template changes.
//
// It uses golden file as reference template and when changes to template are
// intentional, they can be updated by providing -update flag for go test.
//
//  go test ./service/controller/clusterapi/v30/cloudconfig -run Test_Controller_CloudConfig_TCCP_Template_Render -update
//
func Test_Controller_CloudConfig_TCCP_Template_Render(t *testing.T) {
	testCases := []struct {
		name  string
		ctx   context.Context
		cr    v1alpha1.Cluster
		certs certs.Cluster
		keys  randomkeys.Cluster
	}{
		{
			name:  "case 0: tccp test",
			ctx:   unittest.DefaultContext(),
			cr:    unittest.DefaultCluster(),
			certs: unittest.DefaultCerts(),
			keys:  unittest.DefaultKeys(),
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			var cloudConfig *CloudConfig
			{
				ignitionPath, err := k8scloudconfig.GetPackagePath()
				if err != nil {
					t.Fatal(err)
				}

				c := Config{
					Encrypter: &encrypter.EncrypterMock{},
					Logger:    microloggertest.New(),

					CalicoCIDR:                18,
					CalicoMTU:                 1430,
					CalicoSubnet:              "172.18.128.0",
					ClusterIPRange:            "172.18.192.0/22",
					DockerDaemonCIDR:          "172.18.224.1/19",
					IgnitionPath:              ignitionPath,
					ImagePullProgressDeadline: "1m",
					NetworkSetupDockerImage:   "quay.io/giantswarm/k8s-setup-network-environment",
					RegistryDomain:            "quay.io",
					SSHUserList:               "user:ssh-rsa base64==",
					SSOPublicKey:              "user:ssh-rsa base64==",
				}

				cloudConfig, err = New(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			templateBody, err := cloudConfig.NewMasterTemplate(tc.ctx, tc.cr, tc.certs, tc.keys)
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

			if !bytes.Equal([]byte(templateBody), goldenFile) {
				t.Fatalf("\n\n%s\n", cmp.Diff(string(goldenFile), templateBody))
			}
		})
	}
}
