package template

import (
	"strings"
	"testing"
)

func Test_Controller_Resource_CPF_Template_Render(t *testing.T) {
	var err error

	var params *ParamsMain
	{
		recordSets := &ParamsMainRecordSets{
			BaseDomain:     "BaseDomain",
			Route53Enabled: true,
		}
		routeTables := &ParamsMainRouteTables{
			PrivateRoutes: []ParamsMainRouteTablesRoute{
				{
					PeerConnectionID: "PeerConnectionID",
				},
			},
			PublicRoutes: []ParamsMainRouteTablesRoute{
				{
					PeerConnectionID: "PeerConnectionID",
				},
			},
		}

		params = &ParamsMain{
			RecordSets:  recordSets,
			RouteTables: routeTables,
		}
	}

	var templateBody string
	{
		templateBody, err = Render(params)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	{
		expected := "HostedZoneName: 'BaseDomain.'"
		if !strings.Contains(templateBody, expected) {
			t.Fatal("expected", "match", "got", "none")
		}
	}

	{
		expected := "PrivateRoute0:"
		if !strings.Contains(templateBody, expected) {
			t.Fatal("expected", "match", "got", "none")
		}
	}

	{
		expected := "PublicRoute0:"
		if !strings.Contains(templateBody, expected) {
			t.Fatal("expected", "match", "got", "none")
		}
	}
}
