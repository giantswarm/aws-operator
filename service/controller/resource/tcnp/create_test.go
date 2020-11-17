package tcnp

import (
	"bytes"
	"flag"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/ghodss/yaml"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/google/go-cmp/cmp"

	"github.com/giantswarm/aws-operator/service/controller/resource/tcnp/template"
	"github.com/giantswarm/aws-operator/service/internal/changedetection"
	"github.com/giantswarm/aws-operator/service/internal/images"
	"github.com/giantswarm/aws-operator/service/internal/recorder"
	"github.com/giantswarm/aws-operator/service/internal/releases"
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
		cr   infrastructurev1alpha2.AWSMachineDeployment
		re   releasev1alpha1.Release
	}{
		{
			name: "case 0: basic test",
			cr:   unittest.DefaultMachineDeployment(),
			re:   unittest.DefaultRelease(),
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			ctx := unittest.DefaultContext()
			k := unittest.FakeK8sClient()

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

			var e recorder.Interface
			{
				c := recorder.Config{
					K8sClient: k,

					Component: "dummy",
				}

				e = recorder.New(c)
			}

			var d *changedetection.TCNP
			{
				c := changedetection.TCNPConfig{
					Event:    e,
					Logger:   microloggertest.New(),
					Releases: rel,
				}

				d, err = changedetection.NewTCNP(c)
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

			{
				cl := unittest.DefaultCluster()
				err = k.CtrlClient().Create(ctx, &cl)
				if err != nil {
					t.Fatal(err)
				}

				err = k.CtrlClient().Create(ctx, &tc.cr)
				if err != nil {
					t.Fatal(err)
				}

				err = k.CtrlClient().Create(ctx, &tc.re)
				if err != nil {
					t.Fatal(err)
				}
			}

			var r *Resource
			{
				c := Config{
					Detection: d,
					Event:     e,
					Images:    i,
					K8sClient: k,
					Logger:    microloggertest.New(),

					AlikeInstances:   `{"m5.2xlarge":[{"InstanceType":"m5.2xlarge","WeightedCapacity":1},{"InstanceType":"m4.2xlarge","WeightedCapacity":1}]}`,
					InstallationName: "dummy",
				}

				r, err = New(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			params, err := r.newTemplateParams(ctx, tc.cr)
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
