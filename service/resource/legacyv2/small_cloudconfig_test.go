package legacyv2

import (
	"testing"

	"github.com/giantswarm/aws-operator/service/resource/cloudformationv2/adapter"
)

func TestAdapterSmallCloudConfig(t *testing.T) {
	cloudconfigConfig := adapter.SmallCloudconfigConfig{
		MachineType: "machine",
		Region:      "region",
		S3URI:       "s3uri",
	}

	_, err := adapter.SmallCloudconfig(cloudconfigConfig)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
}
