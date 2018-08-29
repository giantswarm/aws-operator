package adapter

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/v16/key"
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
			a := &GuestInstanceAdapter{}
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
			Description:  "case 0 HTTP URL",
			ExpectedLine: fmt.Sprintf("https://s3.eu-west-1.amazonaws.com/000000000000-g8s-test-cluster/version/0.1.0/cloudconfig/%s/master", key.CloudConfigVersion),
			Region:       "eu-west-1",
		},
		{
			Description:  "case 1 S3 URL different region",
			ExpectedLine: fmt.Sprintf("s3://000000000000-g8s-test-cluster/version/0.1.0/cloudconfig/%s/master", key.CloudConfigVersion),
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
					VersionBundle: v1alpha1.AWSConfigSpecVersionBundle{
						Version: "0.1.0",
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

			a := &GuestInstanceAdapter{}
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
