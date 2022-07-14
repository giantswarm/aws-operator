package tccpn

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/ghodss/yaml"
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/google/go-cmp/cmp"

	"github.com/giantswarm/aws-operator/v12/service/controller/resource/tccpn/template"
	"github.com/giantswarm/aws-operator/v12/service/internal/changedetection"
	"github.com/giantswarm/aws-operator/v12/service/internal/cloudtags"
	"github.com/giantswarm/aws-operator/v12/service/internal/encrypter"
	"github.com/giantswarm/aws-operator/v12/service/internal/hamaster"
	"github.com/giantswarm/aws-operator/v12/service/internal/images"
	"github.com/giantswarm/aws-operator/v12/service/internal/recorder"
	"github.com/giantswarm/aws-operator/v12/service/internal/releases"
	"github.com/giantswarm/aws-operator/v12/service/internal/unittest"
)

var update = flag.Bool("update", false, "update .golden CF template file")

// Test_Controller_Resource_TCCPN_Template_Render tests tenant cluster
// CloudFormation template rendering. It is meant to be used as a tool to easily
// check resulting CF template and prevent from accidental CF template changes.
//
// It uses golden file as reference template and when changes to template are
// intentional, they can be updated by providing -update flag for go test.
//
//  go test ./service/controller/resource/tccpn -run Test_Controller_Resource_TCCPN_Template_Render -update
//
func Test_Controller_Resource_TCCPN_Template_Render(t *testing.T) {
	testCases := []struct {
		name           string
		azs            []string
		replicas       int
		route53Enabled bool
		annotations    map[string]string
	}{
		{
			name:           "case 0: basic test with encrypter backend KMS, route53 enabled",
			azs:            []string{"eu-central-1b"},
			replicas:       1,
			route53Enabled: true,
		},
		{
			name:           "case 1: basic test with encrypter backend KMS, route53 disabled",
			azs:            []string{"eu-central-1b"},
			replicas:       1,
			route53Enabled: false,
		},
		{
			name:           "case 2: basic test with encrypter backend KMS, ha masters",
			azs:            []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"},
			replicas:       3,
			route53Enabled: true,
		},
		{
			name:           "case 3: basic test with ebs volume iops and throughput set",
			azs:            []string{"eu-central-1b"},
			replicas:       1,
			route53Enabled: true,
			annotations:    map[string]string{annotation.AWSEBSVolumeIops: "16000", annotation.AWSEBSVolumeThroughput: "1000"},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			ctx := unittest.DefaultContextControlPlane()
			k := unittest.FakeK8sClient()

			var ct cloudtags.Interface
			{
				c := cloudtags.Config{
					K8sClient: k,
					Logger:    microloggertest.New(),
				}

				ct, err = cloudtags.New(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			var m encrypter.Interface
			{
				m = &encrypter.Mock{}
			}

			var e recorder.Interface
			{
				c := recorder.Config{
					K8sClient: k,

					Component: "dummy",
				}

				e = recorder.New(c)
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

			var rel releases.Interface
			{
				c := releases.Config{
					K8sClient: k,
				}

				rel, err = releases.New(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			var d *changedetection.TCCPN
			{
				c := changedetection.TCCPNConfig{
					Event:    e,
					HAMaster: h,
					Logger:   microloggertest.New(),
					Releases: rel,
				}

				d, err = changedetection.NewTCCPN(c)
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

			var aws infrastructurev1alpha3.AWSControlPlane
			{
				cl := unittest.DefaultCluster()
				err = k.CtrlClient().Create(ctx, &cl)
				if err != nil {
					t.Fatal(err)
				}

				aws = unittest.DefaultAWSControlPlane()
				aws.Spec.AvailabilityZones = tc.azs
				for k, v := range tc.annotations {
					fmt.Println(k)
					fmt.Println(v)
					aws.Annotations[k] = v
				}
				err = k.CtrlClient().Create(ctx, &aws)
				if err != nil {
					t.Fatal(err)
				}

				g8s := unittest.DefaultG8sControlPlane()
				g8s.Spec.Replicas = tc.replicas
				err = k.CtrlClient().Create(ctx, &g8s)
				if err != nil {
					t.Fatal(err)
				}

				re := unittest.DefaultRelease()
				err = k.CtrlClient().Create(ctx, &re)
				if err != nil {
					t.Fatal(err)
				}
			}

			var r *Resource
			{
				c := Config{
					CloudTags: ct,
					Encrypter: m,
					Event:     e,
					K8sClient: k,
					Detection: d,
					HAMaster:  h,
					Images:    i,
					Logger:    microloggertest.New(),

					InstallationName: "dummy",
					Route53Enabled:   tc.route53Enabled,
				}

				r, err = New(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			params, err := r.newTemplateParams(ctx, aws)
			if err != nil {
				t.Fatal(err)
			}

			templateBody, err := template.Render(params)
			if err != nil {
				t.Fatal(err)
			}

			_, err = yaml.YAMLToJSONStrict([]byte(templateBody))
			if err != nil {
				t.Fatal(err)
			}

			p := filepath.Join("testdata", unittest.NormalizeFileName(tc.name)+".golden")

			if *update {
				err := ioutil.WriteFile(p, []byte(templateBody), 0644) // nolint: gosec
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
