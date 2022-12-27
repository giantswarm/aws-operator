package tccpf

import (
	"bytes"
	"context"
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/ghodss/yaml"
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/google/go-cmp/cmp"

	"github.com/giantswarm/aws-operator/v14/service/controller/resource/tccpf/template"
	"github.com/giantswarm/aws-operator/v14/service/internal/changedetection"
	"github.com/giantswarm/aws-operator/v14/service/internal/cphostedzone"
	"github.com/giantswarm/aws-operator/v14/service/internal/recorder"
	"github.com/giantswarm/aws-operator/v14/service/internal/unittest"
)

var update = flag.Bool("update", false, "update .golden CF template file")

// Test_Controller_Resource_TCCPF_Template_Render tests tenant cluster
// CloudFormation template rendering. It is meant to be used as a tool to easily
// check resulting CF template and prevent from accidental CF template changes.
//
// It uses golden file as reference template and when changes to template are
// intentional, they can be updated by providing -update flag for go test.
//
//	go test ./service/controller/resource/tccpf -run Test_Controller_Resource_TCCPF_Template_Render -update
func Test_Controller_Resource_TCCPF_Template_Render(t *testing.T) {
	testCases := []struct {
		name           string
		ctx            context.Context
		cr             infrastructurev1alpha3.AWSCluster
		route53Enabled bool
	}{
		{
			name:           "case 0: basic test",
			ctx:            unittest.DefaultContext(),
			cr:             unittest.DefaultCluster(),
			route53Enabled: true,
		},
		{
			name:           "case 1: without route 53 enabled",
			ctx:            unittest.DefaultContext(),
			cr:             unittest.DefaultCluster(),
			route53Enabled: false,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			k := unittest.FakeK8sClient()

			var d *changedetection.TCCPF
			{
				c := changedetection.TCCPFConfig{
					Logger: microloggertest.New(),
				}

				d, err = changedetection.NewTCCPF(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			var e recorder.Interface
			{
				c := recorder.Config{
					K8sClient: k,

					Component: "dummy",
				}

				e = recorder.New(c)
			}

			var h *cphostedzone.HostedZone
			{
				c := cphostedzone.Config{
					Logger: microloggertest.New(),

					Route53Enabled: false,
				}

				h, err = cphostedzone.New(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			var r *Resource
			{
				c := Config{
					Detection:  d,
					Event:      e,
					HostedZone: h,
					Logger:     microloggertest.New(),

					InstallationName: "dummy",
					Route53Enabled:   tc.route53Enabled,
				}

				r, err = New(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			params, err := r.newTemplateParams(tc.ctx, tc.cr)
			if err != nil {
				if err != nil {
					t.Fatal(err)
				}
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
				err := os.WriteFile(p, []byte(templateBody), 0644) // nolint: gosec
				if err != nil {
					t.Fatal(err)
				}
			}
			goldenFile, err := os.ReadFile(p)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal([]byte(templateBody), goldenFile) {
				t.Fatalf("\n\n%s\n", cmp.Diff(string(goldenFile), templateBody))
			}
		})
	}
}
