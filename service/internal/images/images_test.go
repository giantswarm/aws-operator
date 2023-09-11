package images

import (
	"context"
	"os"
	"strconv"
	"testing"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/operatorkit/v7/pkg/controller/context/cachekeycontext"
	releasev1alpha1 "github.com/giantswarm/release-operator/v4/api/v1alpha1"

	"github.com/giantswarm/aws-operator/v14/service/internal/unittest"
)

func Test_Images_Cache(t *testing.T) {
	testCases := []struct {
		name          string
		ctx           context.Context
		expectCaching bool
	}{
		{
			name:          "case 0",
			ctx:           cachekeycontext.NewContext(context.Background(), "1"),
			expectCaching: true,
		},
		{
			name:          "case 1",
			ctx:           context.Background(),
			expectCaching: false,
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
			var ami1 string
			var ami2 string

			var im *Images
			{
				c := Config{
					K8sClient: unittest.FakeK8sClient(),

					RegistryDomain: "dummy",
				}

				im, err = New(c)
				if err != nil {
					t.Fatal(err)
				}
			}

			var cl infrastructurev1alpha3.AWSCluster
			{
				cl = unittest.DefaultCluster()
			}

			var re releasev1alpha1.Release
			{
				re = unittest.DefaultRelease()
			}

			{
				cl.Spec.Provider.Region = "eu-central-1"
				err = im.k8sClient.CtrlClient().Create(tc.ctx, &cl)
				if err != nil {
					t.Fatal(err)
				}

				re.Spec.Components = []releasev1alpha1.ReleaseSpecComponent{
					{
						Name:    "containerlinux",
						Version: "2345.3.0",
					},
				}
				err = im.k8sClient.CtrlClient().Create(tc.ctx, &re)
				if err != nil {
					t.Fatal(err)
				}
			}

			{
				ami1, err = im.AMI(tc.ctx, &cl, "")
				if err != nil {
					t.Fatal(err)
				}
				t.Log(ami1)
			}

			{
				cl.Spec.Provider.Region = "eu-west-1"
				err = im.k8sClient.CtrlClient().Update(tc.ctx, &cl)
				if err != nil {
					t.Fatal(err)
				}

				re.Spec.Components = []releasev1alpha1.ReleaseSpecComponent{
					{
						Name:    "containerlinux",
						Version: "2345.3.1",
					},
				}
				err = im.k8sClient.CtrlClient().Update(tc.ctx, &re)
				if err != nil {
					t.Fatal(err)
				}
			}

			{
				ami2, err = im.AMI(tc.ctx, &cl, "")
				if err != nil {
					t.Fatal(err)
				}
			}

			if tc.expectCaching {
				if ami1 != ami2 {
					t.Fatalf("expected %#q to be equal to %#q", ami1, ami2)
				}
			} else {
				if ami1 == ami2 {
					t.Fatalf("expected %#q to differ from %#q", ami1, ami2)
				}
			}
		})
	}
}
