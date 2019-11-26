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

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset/fake"

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
		name         string
		cr           infrastructurev1alpha2.Cluster
		ctx          context.Context
		errorMatcher func(error) bool
	}{
		{
			name:         "case 0: basic test",
			cr:           unittest.DefaultCluster(),
			ctx:          unittest.DefaultContext(),
			errorMatcher: nil,
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

	var r *Resource
	{

		c := Config{
			CMAClient: fake.NewSimpleClientset(),
			Detection: d,
			Logger:    microloggertest.New(),

			EncrypterBackend: "kms",
		}

		r, err = New(c)
		if err != nil {
			t.Fatal(err)
		}
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			params, err := r.newTemplateParams(tc.ctx, tc.cr, time.Time{})
			if err != nil {
				t.Fatal(err)
			}
			templateBody, err := template.Render(params)
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
