package tccp

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/detection"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/unittest"
)

var update = flag.Bool("update", false, "update .golden CF template file")

// Test_newTemplateBody tests tenant cluster CloudFormation template rendering.
// It is meant to be used as a tool to easily check resulting CF template and
// prevent from accidental CF template changes.
//
// It uses golden file as reference template and when changes to template are
// intentional, they can be updated by providing -update flag for go test.
//
//  go test ./service/controller/clusterapi/v29/resource/tccp -run Test_newTemplateBody -update
//
func Test_newTemplateBody(t *testing.T) {
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

	var d *detection.Detection
	{
		c := detection.Config{
			Logger: microloggertest.New(),
		}

		d, err = detection.New(c)
		if err != nil {
			t.Fatal(err)
		}
	}

	var r *Resource
	{

		c := Config{
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
			templateBody, err := r.newTemplateBody(tc.ctx, tc.cr, tc.tp)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
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
				t.Fatalf("\n\n%s\n", cmp.Diff(templateBody, string(goldenFile)))
			}
		})
	}
}
