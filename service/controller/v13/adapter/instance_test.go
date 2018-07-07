package adapter

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func Test_Adapter_Instance_RegularFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Description              string
		Config                   Config
		ExpectedAZ               string
		ExpectedEtcdVolumeName   string
		ExpectedInstanceType     string
		ExpectedEncrypterBackend string
	}{
		{
			Description: "case 0 basic matching, all fields present",
			Config: Config{
				Clients: Clients{
					EC2: &EC2ClientMock{},
					IAM: &IAMClientMock{},
					STS: &STSClientMock{},
				},
				CustomObject: v1alpha1.AWSConfig{
					Spec: v1alpha1.AWSConfigSpec{
						Cluster: v1alpha1.Cluster{
							ID: "test-cluster",
						},
						AWS: v1alpha1.AWSConfigSpecAWS{
							AZ:     "eu-central-1a",
							Region: "eu-west-1",
						},
					},
				},
				StackState: StackState{
					MasterInstanceType: "m3.large",
				},
				EncrypterBackend: "my-encrypter-backend",
			},
			ExpectedAZ:               "eu-central-1a",
			ExpectedEtcdVolumeName:   "test-cluster-etcd",
			ExpectedInstanceType:     "m3.large",
			ExpectedEncrypterBackend: "my-encrypter-backend",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			a := &instanceAdapter{}
			err := a.Adapt(tc.Config)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			if a.Master.AZ != tc.ExpectedAZ {
				t.Fatalf("unexpected a.Master.AZ, got %q, want %q", a.Master.AZ, tc.ExpectedAZ)
			}

			if a.Master.EtcdVolume.Name != tc.ExpectedEtcdVolumeName {
				t.Fatalf("unexpected a.Master.EtcdVolume.Name, got %q, want %q", a.Master.EtcdVolume.Name, tc.ExpectedEtcdVolumeName)
			}

			if a.Master.Instance.Type != tc.ExpectedInstanceType {
				t.Fatalf("unexpected a.Master.Instance.Type, got %q, want %q", a.Master.Instance.Type, tc.ExpectedInstanceType)
			}

			if a.Master.EncrypterBackend != tc.ExpectedEncrypterBackend {
				t.Fatalf("unexpected a.Master.Instance.Type, got %q, want %q", a.Master.EncrypterBackend, tc.ExpectedEncrypterBackend)
			}
		})
	}
}

func Test_Adapter_Instance_SmallCloudConfig(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Description  string
		ExpectedLine string
		Region       string
	}{
		{
			Description:  "case 0 userdata file",
			ExpectedLine: "USERDATA_FILE=master",
			Region:       "eu-west-1",
		},
		{
			Description:  "scase 1 http URI",
			ExpectedLine: "s3_http_uri=\"https://s3.eu-west-1.amazonaws.com/000000000000-g8s-test-cluster/cloudconfig/foo/$USERDATA_FILE\"",
			Region:       "eu-west-1",
		},
		{
			Description:  "scase 2 http URI different region",
			ExpectedLine: "s3_http_uri=\"https://s3.eu-central-1.amazonaws.com/000000000000-g8s-test-cluster/cloudconfig/foo/$USERDATA_FILE\"",
			Region:       "eu-central-1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			clients := Clients{
				EC2: &EC2ClientMock{},
				IAM: &IAMClientMock{},
				STS: &STSClientMock{accountID: "000000000000"},
			}
			customObject := v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
					},
					AWS: v1alpha1.AWSConfigSpecAWS{
						Masters: []v1alpha1.AWSConfigSpecAWSNode{
							{
								ImageID:      "ami-test",
								InstanceType: "m3.large",
							},
						},
						Region: tc.Region,
					},
				},
			}
			cfg := Config{
				Clients:      clients,
				CustomObject: customObject,
				StackState: StackState{
					MasterCloudConfigVersion: "foo",
				},
			}

			a := &instanceAdapter{}
			err := a.Adapt(cfg)
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			data, err := base64.StdEncoding.DecodeString(a.Master.CloudConfig)
			if err != nil {
				t.Fatalf("unexpected error decoding a.Master.CloudConfig %v", err)
			}

			if !strings.Contains(string(data), tc.ExpectedLine) {
				t.Fatalf("SmallCloudConfig didn't contain expected %q, complete: %q", tc.ExpectedLine, string(data))
			}
		})
	}
}
