package images

import (
	"context"
	"strconv"
	"testing"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/operatorkit/v7/pkg/controller/context/cachekeycontext"
	releasev1alpha1 "github.com/giantswarm/release-operator/v3/api/v1alpha1"

	"github.com/giantswarm/aws-operator/service/internal/unittest"
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
			expectCaching: false,
		},
		{
			name:          "case 1",
			ctx:           context.Background(),
			expectCaching: false,
		},
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
				ami1, err = im.AMI(tc.ctx, &cl)
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
				ami2, err = im.AMI(tc.ctx, &cl)
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
