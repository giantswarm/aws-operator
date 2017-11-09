package certificatetpr

import (
	"io/ioutil"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

func TestSpecYamlEncoding(t *testing.T) {

	spec := Spec{
		AllowBareDomains: true,
		AltNames: []string{
			"kubernetes",
			"kubernetes.default",
			"kubernetes.default.svc",
			"kubernetes.default.svc.cluster.local",
		},
		ClusterComponent: "api",
		ClusterID:        "0yxr7",
		CommonName:       "api.giantswarm.io",
		IPSANs: []string{
			"172.31.0.1",
		},
		Organizations: []string{
			"system:masters",
		},
		TTL: "4320h",
	}

	var got map[string]interface{}
	{
		bytes, err := yaml.Marshal(&spec)
		require.NoError(t, err, "marshaling spec")
		err = yaml.Unmarshal(bytes, &got)
		require.NoError(t, err, "unmarshaling spec to map")
	}

	var want map[string]interface{}
	{
		bytes, err := ioutil.ReadFile("testdata/spec.yaml")
		require.NoError(t, err)
		err = yaml.Unmarshal(bytes, &want)
		require.NoError(t, err, "unmarshaling fixture to map")
	}

	diff := pretty.Compare(want, got)
	require.Equal(t, "", diff, "diff: (-want +got)\n%s", diff)
}
