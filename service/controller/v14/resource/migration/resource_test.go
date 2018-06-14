package migration

import "testing"

func Test_zoneFromAPIDomain(t *testing.T) {
	testCases := []struct {
		name         string
		apiDomain    string
		expectedZone string
		errorMatcher func(err error) bool
	}{
		{
			name:         "case 0: normal case",
			apiDomain:    "api.eggs2.k8s.gauss.eu-central-1.aws.gigantic.io",
			expectedZone: "gauss.eu-central-1.aws.gigantic.io",
		},
		{
			name:         "case 1: domain too short",
			apiDomain:    "api.eggs2.k8s.gigantic",
			errorMatcher: IsMalformedDomain,
		},
		{
			name:         "case 1: minimal length domain",
			apiDomain:    "api.eggs2.k8s.gigantic.io",
			expectedZone: "gigantic.io",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			zone, err := zoneFromAPIDomain(tc.apiDomain)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			if tc.errorMatcher != nil {
				return
			}

			if zone != tc.expectedZone {
				t.Fatalf("zone == %q, want %q", zone, tc.expectedZone)
			}
		})
	}
}
