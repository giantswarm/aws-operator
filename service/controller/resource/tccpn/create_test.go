package tccpn

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/ghodss/yaml"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/google/go-cmp/cmp"

	"github.com/giantswarm/aws-operator/service/controller/resource/tccpn/template"
	"github.com/giantswarm/aws-operator/service/internal/changedetection"
	"github.com/giantswarm/aws-operator/service/internal/hamaster"
	"github.com/giantswarm/aws-operator/service/internal/unittest"
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
		ctx            context.Context
		azs            []string
		replicas       int
		route53Enabled bool
	}{
		{
			name:           "case 0: basic test with encrypter backend KMS, route53 enabled",
			ctx:            unittest.DefaultContextControlPlane(),
			azs:            []string{"eu-central-1a"},
			replicas:       1,
			route53Enabled: true,
		},
		{
			name:           "case 1: basic test with encrypter backend KMS, route53 disabled",
			ctx:            unittest.DefaultContextControlPlane(),
			azs:            []string{"eu-central-1a"},
			replicas:       1,
			route53Enabled: false,
		},
		{
			name:           "case 2: basic test with encrypter backend KMS, ha masters",
			ctx:            unittest.DefaultContextControlPlane(),
			azs:            []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"},
			replicas:       3,
			route53Enabled: true,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			ctx := unittest.DefaultContext()
			k := unittest.FakeK8sClient()

			var d *changedetection.TCCPN
			{
				c := changedetection.TCCPNConfig{
					Logger: microloggertest.New(),
				}

				d, err = changedetection.NewTCCPN(c)
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

			var aws infrastructurev1alpha2.AWSControlPlane
			{
				cl := unittest.DefaultCluster()
				err = k.CtrlClient().Create(ctx, &cl)
				if err != nil {
					t.Fatal(err)
				}

				aws = unittest.DefaultAWSControlPlane()
				aws.Spec.AvailabilityZones = tc.azs
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
			}

			var r *Resource
			{
				c := Config{
					K8sClient: k,
					Detection: d,
					HAMaster:  h,
					Logger:    microloggertest.New(),

					InstallationName: "dummy",
					Route53Enabled:   tc.route53Enabled,
				}

				r, err = New(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			params, err := r.newTemplateParams(tc.ctx, aws)
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
