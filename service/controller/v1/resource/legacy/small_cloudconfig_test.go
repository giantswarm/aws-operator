package legacy

import (
	"testing"

	"github.com/giantswarm/aws-operator/service/controller/v2/resource/cloudformation/adapter"
)

func TestAdapterSmallCloudConfig(t *testing.T) {
	t.Parallel()
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
