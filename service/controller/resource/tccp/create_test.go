package tccp

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/ghodss/yaml"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/google/go-cmp/cmp"

	"github.com/giantswarm/aws-operator/service/controller/internal/changedetection"
	"github.com/giantswarm/aws-operator/service/controller/internal/unittest"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccp/template"
)

var update = flag.Bool("update", false, "update .golden CF template file")

// Test_Controller_Resource_TCCP_Template_Render tests tenant cluster
// CloudFormation template rendering. It is meant to be used as a tool to easily
// check resulting CF template and prevent from accidental CF template changes.
//
// It uses golden file as reference template and when changes to template are
// intentional, they can be updated by providing -update flag for go test.
//
//  go test ./service/controller/resource/tccp -run Test_Controller_Resource_TCCP_Template_Render -update
//
func Test_Controller_Resource_TCCP_Template_Render(t *testing.T) {
	testCases := []struct {
		name           string
		cr             infrastructurev1alpha2.AWSCluster
		ctx            context.Context
		apiWhitelist   ConfigAPIWhitelistSecurityGroup
		route53Enabled bool
		errorMatcher   func(error) bool
	}{
		{
			name:           "case 0: basic test, route53 enabled",
			cr:             unittest.DefaultCluster(),
			ctx:            unittest.DefaultContext(),
			errorMatcher:   nil,
			route53Enabled: true,
		},
		{
			name:           "case 1: basic test, route53 disabled",
			cr:             unittest.DefaultCluster(),
			ctx:            unittest.DefaultContext(),
			errorMatcher:   nil,
			route53Enabled: false,
		},
		{
			name: "case 2: basic test with api whitelist enabled",
			cr:   unittest.DefaultCluster(),
			ctx:  unittest.DefaultContext(),
			apiWhitelist: ConfigAPIWhitelistSecurityGroup{
				Enabled: true,
				SubnetList: []string{
					"172.10.10.10",
					"172.20.20.20",
				},
			},
			errorMatcher:   nil,
			route53Enabled: false,
		},
	}

	var err error

	var d *changedetection.TCCP
	{
		c := changedetection.TCCPConfig{
			Logger: microloggertest.New(),
		}

		d, err = changedetection.NewTCCP(c)
		if err != nil {
			t.Fatal(err)
		}
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var r *Resource
			{

				c := Config{
					G8sClient: fake.NewSimpleClientset(),
					Detection: d,
					Logger:    microloggertest.New(),

					APIWhitelist: ConfigAPIWhitelist{
						Private: ConfigAPIWhitelistSecurityGroup{
							Enabled:    tc.apiWhitelist.Enabled,
							SubnetList: tc.apiWhitelist.SubnetList,
						},
						Public: ConfigAPIWhitelistSecurityGroup{
							Enabled:    tc.apiWhitelist.Enabled,
							SubnetList: tc.apiWhitelist.SubnetList,
						},
					},
					CIDRBlockAWSCNI: "172.17.0.1/16",
					Route53Enabled:  tc.route53Enabled,
				}

				r, err = New(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			params, err := r.newParamsMain(tc.ctx, tc.cr, time.Time{})
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
