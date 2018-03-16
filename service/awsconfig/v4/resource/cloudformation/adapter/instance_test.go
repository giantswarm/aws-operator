package adapter

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func TestAdapterInstanceRegularFields(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		description             string
		customObject            v1alpha1.AWSConfig
		errorMatcher            func(error) bool
		expectedAZ              string
		expectedInstanceType    string
		expectedSecurityGroupID string
	}{
		{
			description:  "empty custom object",
			customObject: v1alpha1.AWSConfig{},
			errorMatcher: IsInvalidConfig,
		},
		{
			description: "basic matching, all fields present",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
					},
					AWS: v1alpha1.AWSConfigSpecAWS{
						AZ: "eu-central-1a",
						Masters: []v1alpha1.AWSConfigSpecAWSNode{
							v1alpha1.AWSConfigSpecAWSNode{
								InstanceType: "m3.large",
							},
						},
					},
				},
			},
			errorMatcher:         nil,
			expectedAZ:           "eu-central-1a",
			expectedInstanceType: "m3.large",
		},
	}

	for _, tc := range testCases {
		clients := Clients{
			EC2: &EC2ClientMock{},
			IAM: &IAMClientMock{},
		}
		a := Adapter{}

		t.Run(tc.description, func(t *testing.T) {
			cfg := Config{
				CustomObject: tc.customObject,
				Clients:      clients,
			}
			err := a.getInstance(cfg)
			if tc.errorMatcher != nil && err == nil {
				t.Error("expected error didn't happen")
			}

			if tc.errorMatcher != nil && !tc.errorMatcher(err) {
				t.Error("expected", true, "got", false)
			}

			if a.MasterAZ != tc.expectedAZ {
				t.Errorf("unexpected MasterAZ, got %q, want %q", a.instanceAdapter.MasterAZ, tc.expectedAZ)
			}

			if a.MasterInstanceType != tc.expectedInstanceType {
				t.Errorf("unexpected MasterInstanceType, got %q, want %q", a.instanceAdapter.MasterInstanceType, tc.expectedInstanceType)
			}
		})
	}
}

func TestAdapterInstanceSmallCloudConfig(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		description  string
		expectedLine string
	}{
		{
			description:  "userdata file",
			expectedLine: "USERDATA_FILE=master",
		},
		{
			description:  "s3 http uri",
			expectedLine: `s3_http_uri="https://s3.myregion.amazonaws.com/000000000000-g8s-test-cluster/cloudconfig/v_3_1_0/$USERDATA_FILE"`,
		},
	}

	a := Adapter{}
	clients := Clients{
		EC2: &EC2ClientMock{},
		IAM: &IAMClientMock{accountID: "000000000000"},
	}
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
			AWS: v1alpha1.AWSConfigSpecAWS{
				Region: "myregion",
				Masters: []v1alpha1.AWSConfigSpecAWSNode{
					v1alpha1.AWSConfigSpecAWSNode{
						ImageID:      "ami-test",
						InstanceType: "m3.large",
					},
				},
			},
		},
	}
	cfg := Config{
		CustomObject: customObject,
		Clients:      clients,
	}
	err := a.getInstance(cfg)

	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	data, err := base64.StdEncoding.DecodeString(a.MasterSmallCloudConfig)
	if err != nil {
		t.Errorf("unexpected error decoding SmallCloudConfig %v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			if !strings.Contains(string(data), tc.expectedLine) {
				t.Errorf("SmallCloudConfig didn't contain expected %q, complete: %q", tc.expectedLine, string(data))
			}
		})
	}
}
