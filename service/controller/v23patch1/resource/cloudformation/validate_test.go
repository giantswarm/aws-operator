package cloudformation

import (
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-operator/service/controller/v23/adapter"
)

func Test_validateHostPeeringRoutes(t *testing.T) {
	t.Parallel()
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			AWS: v1alpha1.AWSConfigSpecAWS{
				VPC: v1alpha1.AWSConfigSpecAWSVPC{
					PrivateSubnetCIDR: "172.31.0.0/16",
				},
			},
		},
	}

	testCases := []struct {
		description         string
		matchingRouteTables int
		expectedError       bool
	}{
		{
			description:         "route table doesn't exist, do not expect error",
			matchingRouteTables: 0,
			expectedError:       false,
		},
		{
			description:         "route table exists, expect error",
			matchingRouteTables: 1,
			expectedError:       true,
		},
		{
			description:         "two route table exist, expect error",
			matchingRouteTables: 2,
			expectedError:       true,
		},
	}

	for _, tc := range testCases {
		var err error
		var newResource *Resource
		{
			ec2Mock := &adapter.EC2ClientMock{}
			ec2Mock.SetMatchingRouteTables(tc.matchingRouteTables)

			c := Config{}

			c.EncrypterBackend = "kms"
			c.G8sClient = fake.NewSimpleClientset()
			c.HostClients = &adapter.Clients{
				EC2:            ec2Mock,
				CloudFormation: &adapter.CloudFormationMock{},
				IAM:            &adapter.IAMClientMock{},
				STS:            &adapter.STSClientMock{},
			}
			c.Logger = microloggertest.New()

			newResource, err = New(c)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}
		}

		t.Run(tc.description, func(t *testing.T) {
			err := newResource.validateHostPeeringRoutes(customObject)
			if tc.expectedError && err == nil {
				t.Fatalf("expected error didn't happen")
			}
			if !tc.expectedError && err != nil {
				t.Fatalf("unexpected error %v", err)
			}
		})
	}
}
