package cpf

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

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/cpf/template"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/unittest"
)

var update = flag.Bool("update", false, "update .golden CF template file")

// Test_Controller_Resource_CPF_Template_Render tests tenant cluster
// CloudFormation template rendering. It is meant to be used as a tool to easily
// check resulting CF template and prevent from accidental CF template changes.
//
// It uses golden file as reference template and when changes to template are
// intentional, they can be updated by providing -update flag for go test.
//
//  go test ./service/controller/clusterapi/v29/resource/cpf -run Test_Controller_Resource_CPF_Template_Render -update
//
func Test_Controller_Resource_CPF_Template_Render(t *testing.T) {
	testCases := []struct {
		name string
		ctx  context.Context
		cl   v1alpha1.Cluster
	}{
		{
			name: "case 0: basic test",
			ctx:  unittest.DefaultContext(),
			cl:   unittest.DefaultCluster(),
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			var r *Resource
			{
				c := Config{
					Logger: microloggertest.New(),

					EncrypterBackend: encrypter.KMSBackend,
					Route53Enabled:   true,
				}

				r, err = New(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			var params *template.ParamsMain
			{
				recordSets, err := r.newRecordSetsParams(tc.ctx, tc.cl)
				if err != nil {
					t.Fatal(err)
				}
				routeTables, err := r.newRouteTablesParams(tc.ctx, tc.cl)
				if err != nil {
					t.Fatal(err)
				}

				params = &template.ParamsMain{
					RecordSets:  recordSets,
					RouteTables: routeTables,
				}
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
				t.Fatalf("\n\n%s\n", cmp.Diff(templateBody, string(goldenFile)))
			}
		})
	}
}
