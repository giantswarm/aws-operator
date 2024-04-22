package tcnp

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/ghodss/yaml"
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/micrologger/microloggertest"
	releasev1alpha1 "github.com/giantswarm/release-operator/v4/api/v1alpha1"
	"github.com/google/go-cmp/cmp"

	"github.com/giantswarm/aws-operator/v14/service/controller/resource/tcnp/template"
	"github.com/giantswarm/aws-operator/v14/service/internal/changedetection"
	"github.com/giantswarm/aws-operator/v14/service/internal/cloudtags"
	"github.com/giantswarm/aws-operator/v14/service/internal/encrypter"
	"github.com/giantswarm/aws-operator/v14/service/internal/images"
	"github.com/giantswarm/aws-operator/v14/service/internal/recorder"
	"github.com/giantswarm/aws-operator/v14/service/internal/releases"
	"github.com/giantswarm/aws-operator/v14/service/internal/unittest"
)

var update = flag.Bool("update", false, "update .golden CF template file")

// Test_Controller_Resource_TCNP_Template_Render tests tenant cluster
// CloudFormation template rendering. It is meant to be used as a tool to easily
// check resulting CF template and prevent from accidental CF template changes.
//
// It uses golden file as reference template and when changes to template are
// intentional, they can be updated by providing -update flag for go test.
//
//	go test ./service/controller/resource/tcnp -run Test_Controller_Resource_TCNP_Template_Render -update
func Test_Controller_Resource_TCNP_Template_Render(t *testing.T) {
	testCases := []struct {
		name string
		cr   infrastructurev1alpha3.AWSMachineDeployment
		re   releasev1alpha1.Release
	}{
		{
			name: "case 0: basic test",
			cr:   unittest.DefaultMachineDeployment(),
			re:   unittest.DefaultRelease(),
		},
		{
			name: "case 1: disk test",
			cr:   unittest.MachineDeploymentWithDisks(unittest.DefaultMachineDeployment(), "10", 11, 12, "13"),
			re:   unittest.DefaultRelease(),
		},
	}

	data := `{
  "2345.3.0": {
    "ap-east-1": "ami-0a813620447e80b05",
    "ap-northeast-1": "ami-02af6d096f0a2f96c",
    "ap-northeast-2": "ami-0dd7a397031040ff1",
    "ap-south-1": "ami-03205de7c095444bb",
    "ap-southeast-1": "ami-0666b8c77a4148316",
    "ap-southeast-2": "ami-0be9b0ada4e9f0c7a",
    "ca-central-1": "ami-0b0044f3e521384ae",
    "eu-central-1": "ami-0c9ac894c7ec2e6dd",
    "eu-north-1": "ami-06420e2a1713889dd",
    "eu-west-1": "ami-09f0cd6af1e455cd9",
    "eu-west-2": "ami-0b1588137b7790e8c",
    "eu-west-3": "ami-01a8e028daf4d66cf",
    "me-south-1": "ami-0a6518241a90f491f",
    "sa-east-1": "ami-037e10c3bd117fb3e",
    "us-east-1": "ami-007776654941e2586",
    "us-east-2": "ami-0b0a4944bd30c6b85",
    "us-west-1": "ami-02fefca3d52d15b1d",
    "us-west-2": "ami-0390d41fd4e4a3529"
  },
  "2345.3.1": {
    "ap-east-1": "ami-0e28e38ecce552688",
    "ap-northeast-1": "ami-074891de68922e1f4",
    "ap-northeast-2": "ami-0a1a6a05c79bcdfe4",
    "ap-south-1": "ami-0765ae35424be8ad8",
    "ap-southeast-1": "ami-0f20e37280d5c8c5c",
    "ap-southeast-2": "ami-016e5e9a74cc6ef86",
    "ca-central-1": "ami-09afcf2e90761d6e6",
    "cn-north-1": "ami-019174dba14053d2a",
    "cn-northwest-1": "ami-004e81bc53b1e6ffa",
    "eu-central-1": "ami-0a9a5d2b65cce04eb",
    "eu-north-1": "ami-0bbfc19aa4c355fe2",
    "eu-west-1": "ami-002db020452770c0f",
    "eu-west-2": "ami-024928e37dcc18a42",
    "eu-west-3": "ami-083e4a190c9b050b1",
    "me-south-1": "ami-078eb26f287443167",
    "sa-east-1": "ami-01180d594d0315f65",
    "us-east-1": "ami-011655f166912d5ba",
    "us-east-2": "ami-0e30f3d8cbc900ff4",
    "us-west-1": "ami-0360d32ce24f1f05f",
    "us-west-2": "ami-0c1654a9988866a1f"
  }
}`
	err := os.WriteFile("/tmp/ami.json", []byte(data), os.ModePerm) // nolint:gosec
	if err != nil {
		panic(err)
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			ctx := unittest.DefaultContext()
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
				awsCl := unittest.DefaultCluster()
				err = k.CtrlClient().Create(ctx, &awsCl)
				if err != nil {
					t.Fatal(err)
				}

				cl := unittest.DefaultCAPIClusterWithLabels(awsCl.Name, map[string]string{})
				err = k.CtrlClient().Create(ctx, &cl)
				if err != nil {
					t.Fatal(err)
				}

				err = k.CtrlClient().Create(ctx, &tc.cr) // nolint:gosec
				if err != nil {
					t.Fatal(err)
				}

				err = k.CtrlClient().Create(ctx, &tc.re) // nolint:gosec
				if err != nil {
					t.Fatal(err)
				}
			}

			var r *Resource
			{
				c := Config{
					CloudTags: ct,
					Detection: d,
					Encrypter: m,
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

	_ = os.Remove("/tmp/ami.json")
}
