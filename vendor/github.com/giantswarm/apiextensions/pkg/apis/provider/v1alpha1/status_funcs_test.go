package v1alpha1

import (
	"reflect"
	"testing"
	"time"
)

func Test_Provider_Status_LatestVersion(t *testing.T) {
	testCases := []struct {
		Name           string
		StatusCluster  StatusCluster
		ExpectedSemver string
	}{
		{
			Name: "case 0",
			StatusCluster: StatusCluster{
				Versions: []StatusClusterVersion{},
			},
			ExpectedSemver: "",
		},
		{
			Name: "case 1",
			StatusCluster: StatusCluster{
				Versions: []StatusClusterVersion{
					{
						Date:   time.Unix(10, 0),
						Semver: "1.0.0",
					},
				},
			},
			ExpectedSemver: "1.0.0",
		},
		{
			Name: "case 2",
			StatusCluster: StatusCluster{
				Versions: []StatusClusterVersion{
					{
						Date:   time.Unix(10, 0),
						Semver: "1.0.0",
					},
					{
						Date:   time.Unix(20, 0),
						Semver: "2.0.0",
					},
				},
			},
			ExpectedSemver: "2.0.0",
		},
		{
			Name: "case 3",
			StatusCluster: StatusCluster{
				Versions: []StatusClusterVersion{
					{
						Date:   time.Unix(10, 0),
						Semver: "1.0.0",
					},
					{
						Date:   time.Unix(20, 0),
						Semver: "2.0.0",
					},
					{
						Date:   time.Unix(30, 0),
						Semver: "3.0.0",
					},
				},
			},
			ExpectedSemver: "3.0.0",
		},
		{
			Name: "case 4",
			StatusCluster: StatusCluster{
				Versions: []StatusClusterVersion{
					{
						Date:   time.Unix(20, 0),
						Semver: "2.0.0",
					},
					{
						Date:   time.Unix(30, 0),
						Semver: "3.0.0",
					},
					{
						Date:   time.Unix(10, 0),
						Semver: "1.0.0",
					},
				},
			},
			ExpectedSemver: "3.0.0",
		},
		{
			Name: "case 5",
			StatusCluster: StatusCluster{
				Versions: []StatusClusterVersion{
					{
						LastTransitionTime: DeepCopyTime{
							time.Unix(20, 0),
						},
						Date:   time.Time{},
						Semver: "2.0.0",
					},
					{
						LastTransitionTime: DeepCopyTime{
							time.Unix(30, 0),
						},
						Date:   time.Time{},
						Semver: "3.0.0",
					},
					{
						LastTransitionTime: DeepCopyTime{
							time.Unix(10, 0),
						},
						Date:   time.Time{},
						Semver: "1.0.0",
					},
				},
			},
			ExpectedSemver: "3.0.0",
		},
	}

	for _, tc := range testCases {
		semver := tc.StatusCluster.LatestVersion()

		if semver != tc.ExpectedSemver {
			t.Fatalf("expected %#v got %#v", tc.ExpectedSemver, semver)
		}
	}
}

func Test_Provider_Status_withCondition(t *testing.T) {
	testTime := time.Unix(20, 0)

	testCases := []struct {
		Name               string
		Conditions         []StatusClusterCondition
		Search             string
		Replace            string
		Status             string
		ExpectedConditions []StatusClusterCondition
	}{
		{
			Name:       "case 0",
			Conditions: []StatusClusterCondition{},
			Search:     StatusClusterTypeCreating,
			Replace:    StatusClusterTypeCreated,
			Status:     StatusClusterStatusTrue,
			ExpectedConditions: []StatusClusterCondition{
				{
					LastTransitionTime: DeepCopyTime{testTime},
					Status:             StatusClusterStatusTrue,
					Type:               StatusClusterTypeCreated,
				},
			},
		},
		{
			Name: "case 1",
			Conditions: []StatusClusterCondition{
				{
					LastTransitionTime: DeepCopyTime{testTime},
					Status:             StatusClusterStatusTrue,
					Type:               StatusClusterTypeCreating,
				},
			},
			Search:  StatusClusterTypeCreating,
			Replace: StatusClusterTypeCreated,
			Status:  StatusClusterStatusTrue,
			ExpectedConditions: []StatusClusterCondition{
				{
					LastTransitionTime: DeepCopyTime{testTime},
					Status:             StatusClusterStatusTrue,
					Type:               StatusClusterTypeCreated,
				},
			},
		},
	}

	for _, tc := range testCases {
		conditions := withCondition(tc.Conditions, tc.Search, tc.Replace, tc.Status, testTime)

		if !reflect.DeepEqual(conditions, tc.ExpectedConditions) {
			t.Fatalf("%s: expected %#v got %#v", tc.Name, tc.ExpectedConditions, conditions)
		}
	}
}

func Test_Provider_Status_withVersion(t *testing.T) {
	testCases := []struct {
		Name             string
		Versions         []StatusClusterVersion
		Version          StatusClusterVersion
		Limit            int
		ExpectedVersions []StatusClusterVersion
	}{
		{
			Name:     "case 0: list with zero items results in a list with one item",
			Versions: []StatusClusterVersion{},
			Version: StatusClusterVersion{
				Date:   time.Unix(10, 0),
				Semver: "1.0.0",
			},
			Limit: 3,
			ExpectedVersions: []StatusClusterVersion{
				{
					Date:   time.Unix(10, 0),
					Semver: "1.0.0",
				},
			},
		},
		{
			Name: "case 1: list with one item results in a list with two items",
			Versions: []StatusClusterVersion{
				{
					Date:   time.Unix(10, 0),
					Semver: "1.0.0",
				},
			},
			Version: StatusClusterVersion{
				Date:   time.Unix(20, 0),
				Semver: "1.1.0",
			},
			Limit: 3,
			ExpectedVersions: []StatusClusterVersion{
				{
					Date:   time.Unix(10, 0),
					Semver: "1.0.0",
				},
				{
					Date:   time.Unix(20, 0),
					Semver: "1.1.0",
				},
			},
		},
		{
			Name: "case 2: list with two items results in a list with three items",
			Versions: []StatusClusterVersion{
				{
					Date:   time.Unix(10, 0),
					Semver: "1.0.0",
				},
				{
					Date:   time.Unix(20, 0),
					Semver: "1.1.0",
				},
			},
			Version: StatusClusterVersion{
				Date:   time.Unix(30, 0),
				Semver: "1.5.0",
			},
			Limit: 3,
			ExpectedVersions: []StatusClusterVersion{
				{
					Date:   time.Unix(10, 0),
					Semver: "1.0.0",
				},
				{
					Date:   time.Unix(20, 0),
					Semver: "1.1.0",
				},
				{
					Date:   time.Unix(30, 0),
					Semver: "1.5.0",
				},
			},
		},
		{
			Name: "case 3: list with three items results in a list with three items due to limit",
			Versions: []StatusClusterVersion{
				{
					Date:   time.Unix(10, 0),
					Semver: "1.0.0",
				},
				{
					Date:   time.Unix(20, 0),
					Semver: "1.1.0",
				},
				{
					Date:   time.Unix(30, 0),
					Semver: "1.5.0",
				},
			},
			Version: StatusClusterVersion{
				Date:   time.Unix(40, 0),
				Semver: "3.0.0",
			},
			Limit: 3,
			ExpectedVersions: []StatusClusterVersion{
				{
					Date:   time.Unix(20, 0),
					Semver: "1.1.0",
				},
				{
					Date:   time.Unix(30, 0),
					Semver: "1.5.0",
				},
				{
					Date:   time.Unix(40, 0),
					Semver: "3.0.0",
				},
			},
		},
		{
			Name: "case 4: list with five items results in a list with three items due to limit",
			Versions: []StatusClusterVersion{
				{
					Date:   time.Unix(10, 0),
					Semver: "1.0.0",
				},
				{
					Date:   time.Unix(20, 0),
					Semver: "1.1.0",
				},
				{
					Date:   time.Unix(30, 0),
					Semver: "1.5.0",
				},
				{
					Date:   time.Unix(40, 0),
					Semver: "3.0.0",
				},
				{
					Date:   time.Unix(50, 0),
					Semver: "3.2.0",
				},
			},
			Version: StatusClusterVersion{
				Date:   time.Unix(60, 0),
				Semver: "3.3.0",
			},
			Limit: 3,
			ExpectedVersions: []StatusClusterVersion{
				{
					Date:   time.Unix(40, 0),
					Semver: "3.0.0",
				},
				{
					Date:   time.Unix(50, 0),
					Semver: "3.2.0",
				},
				{
					Date:   time.Unix(60, 0),
					Semver: "3.3.0",
				},
			},
		},
		{
			Name: "case 5: same as 4 but checking items are ordered by date before cutting off",
			Versions: []StatusClusterVersion{
				{
					Date:   time.Unix(40, 0),
					Semver: "3.0.0",
				},
				{
					Date:   time.Unix(20, 0),
					Semver: "1.1.0",
				},
				{
					Date:   time.Unix(10, 0),
					Semver: "1.0.0",
				},
				{
					Date:   time.Unix(50, 0),
					Semver: "3.2.0",
				},
				{
					Date:   time.Unix(30, 0),
					Semver: "1.5.0",
				},
			},
			Version: StatusClusterVersion{
				Date:   time.Unix(60, 0),
				Semver: "3.3.0",
			},
			Limit: 3,
			ExpectedVersions: []StatusClusterVersion{
				{
					Date:   time.Unix(40, 0),
					Semver: "3.0.0",
				},
				{
					Date:   time.Unix(50, 0),
					Semver: "3.2.0",
				},
				{
					Date:   time.Unix(60, 0),
					Semver: "3.3.0",
				},
			},
		},
		{
			Name: "case 6: list with one item results in a list with one item in case the version already exists",
			Versions: []StatusClusterVersion{
				{
					Date:   time.Unix(10, 0),
					Semver: "1.0.0",
				},
			},
			Version: StatusClusterVersion{
				Date:   time.Unix(20, 0),
				Semver: "1.0.0",
			},
			Limit: 3,
			ExpectedVersions: []StatusClusterVersion{
				{
					Date:   time.Unix(10, 0),
					Semver: "1.0.0",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			versions := withVersion(tc.Versions, tc.Version, tc.Limit)

			if !reflect.DeepEqual(versions, tc.ExpectedVersions) {
				t.Fatalf("expected %#v got %#v", tc.ExpectedVersions, versions)
			}
		})
	}
}
