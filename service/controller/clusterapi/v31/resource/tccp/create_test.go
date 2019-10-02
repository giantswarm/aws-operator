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

	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset/fake"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/changedetection"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/resource/tccp/template"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/unittest"
)

var update = flag.Bool("update", false, "update .golden CF template file")

// Test_Controller_Resource_TCCP_Template_Render tests tenant cluster
// CloudFormation template rendering. It is meant to be used as a tool to easily
// check resulting CF template and prevent from accidental CF template changes.
//
// It uses golden file as reference template and when changes to template are
// intentional, they can be updated by providing -update flag for go test.
//
//  go test ./service/controller/clusterapi/v31/resource/tccp -run Test_Controller_Resource_TCCP_Template_Render -update
//
func Test_Controller_Resource_TCCP_Template_Render(t *testing.T) {
	testCases := []struct {
		name         string
		cr           v1alpha1.Cluster
		ctx          context.Context
		tp           templateParams
		errorMatcher func(error) bool
	}{
		{
			name: "case 0: basic test",
			cr:   unittest.DefaultCluster(),
			ctx:  unittest.DefaultContext(),
			tp: templateParams{
				DockerVolumeResourceName:   "rsc-abbacd01",
				MasterInstanceResourceName: "rsc-ac0dc01",
			},
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
			VPCPeerID:        "vpc-f8d0e10b",
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
