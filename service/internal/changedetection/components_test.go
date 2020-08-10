package changedetection

import (
	"strconv"
	"testing"

	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/google/go-cmp/cmp"
)

func TestReleaseComponentsEqual(t *testing.T) {
	testCases := []struct {
		name           string
		currentRelease releasev1alpha1.Release
		targetRelease  releasev1alpha1.Release
		result         bool
	}{
		{
			name: "case 0",
			currentRelease: releasev1alpha1.Release{
				Spec: releasev1alpha1.ReleaseSpec{
					Components: []releasev1alpha1.ReleaseSpecComponent{
						{
							Catalog:               "",
							Name:                  "calico",
							Reference:             "",
							ReleaseOperatorDeploy: false,
							Version:               "1.1.0",
						},
					},
				},
			},
			targetRelease: releasev1alpha1.Release{
				Spec: releasev1alpha1.ReleaseSpec{
					Components: []releasev1alpha1.ReleaseSpecComponent{
						{
							Catalog:               "",
							Name:                  "calico",
							Reference:             "",
							ReleaseOperatorDeploy: false,
							Version:               "1.1.1",
						},
					},
				},
			},
			result: false,
		},
		{
			name: "case 1",
			currentRelease: releasev1alpha1.Release{
				Spec: releasev1alpha1.ReleaseSpec{
					Components: []releasev1alpha1.ReleaseSpecComponent{
						{
							Catalog:               "",
							Name:                  "kubernetes",
							Reference:             "",
							ReleaseOperatorDeploy: false,
							Version:               "1.16.9",
						},
						{
							Catalog:               "",
							Name:                  "aws-cni",
							Reference:             "",
							ReleaseOperatorDeploy: false,
							Version:               "1.7.0",
						},
					},
				},
			},
			targetRelease: releasev1alpha1.Release{
				Spec: releasev1alpha1.ReleaseSpec{
					Components: []releasev1alpha1.ReleaseSpecComponent{
						{
							Catalog:               "",
							Name:                  "kubernetes",
							Reference:             "",
							ReleaseOperatorDeploy: false,
							Version:               "1.17.11",
						},
						{
							Catalog:               "",
							Name:                  "calico",
							Reference:             "",
							ReleaseOperatorDeploy: false,
							Version:               "1.1.1",
						},
					},
				},
			},
			result: false,
		},
		{
			name:           "case 2",
			currentRelease: releasev1alpha1.Release{},
			targetRelease: releasev1alpha1.Release{
				Spec: releasev1alpha1.ReleaseSpec{
					Components: []releasev1alpha1.ReleaseSpecComponent{
						{
							Catalog:               "",
							Name:                  "kubernetes",
							Reference:             "",
							ReleaseOperatorDeploy: false,
							Version:               "1.17.11",
						},
						{
							Catalog:               "",
							Name:                  "calico",
							Reference:             "",
							ReleaseOperatorDeploy: false,
							Version:               "1.1.1",
						},
					},
				},
			},
			result: false,
		},
		{
			name: "case 3",
			currentRelease: releasev1alpha1.Release{
				Spec: releasev1alpha1.ReleaseSpec{
					Components: []releasev1alpha1.ReleaseSpecComponent{
						{
							Catalog:               "",
							Name:                  "kubernetes",
							Reference:             "",
							ReleaseOperatorDeploy: false,
							Version:               "1.16.9",
						},
						{
							Catalog:               "",
							Name:                  "aws-cni",
							Reference:             "",
							ReleaseOperatorDeploy: false,
							Version:               "1.7.0",
						},
					},
				},
			},
			targetRelease: releasev1alpha1.Release{
				Spec: releasev1alpha1.ReleaseSpec{
					Components: []releasev1alpha1.ReleaseSpecComponent{
						{
							Catalog:               "",
							Name:                  "kubernetes",
							Reference:             "",
							ReleaseOperatorDeploy: false,
							Version:               "1.16.9",
						},
						{
							Catalog:               "",
							Name:                  "calico",
							Reference:             "",
							ReleaseOperatorDeploy: false,
							Version:               "1.7.0",
						},
					},
				},
			},
			result: true,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result := releaseComponentsEqual(tc.currentRelease, tc.targetRelease)

			if result != tc.result {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.result, result))
			}
		})
	}

}
