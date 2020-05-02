package tcnp

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
	"github.com/google/go-cmp/cmp"

	"github.com/giantswarm/aws-operator/service/controller/resource/tcnp/template"
	"github.com/giantswarm/aws-operator/service/internal/unittest"
)

var update = flag.Bool("update", false, "update .golden CF template file")

// Test_Controller_Resource_TCNP_Template_Render tests tenant cluster
// CloudFormation template rendering. It is meant to be used as a tool to easily
// check resulting CF template and prevent from accidental CF template changes.
//
// It uses golden file as reference template and when changes to template are
// intentional, they can be updated by providing -update flag for go test.
//
//  go test ./service/controller/resource/tcnp -run Test_Controller_Resource_TCNP_Template_Render -update
//
func Test_Controller_Resource_TCNP_Template_Render(t *testing.T) {
	testCases := []struct {
		name string
		ctx  context.Context
		cr   infrastructurev1alpha2.AWSMachineDeployment
	}{
		{
			name: "case 0: basic test",
			ctx:  unittest.DefaultContext(),
			cr:   unittest.DefaultMachineDeployment(),
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			params, err := newTemplateParams(tc.ctx, tc.cr)
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
