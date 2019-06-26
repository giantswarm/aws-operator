package v1alpha1

import (
	"sort"
	"time"
)

func (s CommonClusterStatus) GetCreatedCondition() CommonClusterStatusCondition {
	return getCondition(s.Conditions, ClusterStatusConditionCreated)
}

func (s CommonClusterStatus) GetCreatingCondition() CommonClusterStatusCondition {
	return getCondition(s.Conditions, ClusterStatusConditionCreating)
}

func (s CommonClusterStatus) GetDeletedCondition() CommonClusterStatusCondition {
	return getCondition(s.Conditions, ClusterStatusConditionDeleted)
}

func (s CommonClusterStatus) GetDeletingCondition() CommonClusterStatusCondition {
	return getCondition(s.Conditions, ClusterStatusConditionDeleting)
}

func (s CommonClusterStatus) GetUpdatedCondition() CommonClusterStatusCondition {
	return getCondition(s.Conditions, ClusterStatusConditionUpdated)
}

func (s CommonClusterStatus) GetUpdatingCondition() CommonClusterStatusCondition {
	return getCondition(s.Conditions, ClusterStatusConditionUpdating)
}

func (s CommonClusterStatus) HasCreatedCondition() bool {
	return hasCondition(s.Conditions, ClusterStatusConditionCreated)
}

func (s CommonClusterStatus) HasCreatingCondition() bool {
	return hasCondition(s.Conditions, ClusterStatusConditionCreating)
}

func (s CommonClusterStatus) HasDeletedCondition() bool {
	return hasCondition(s.Conditions, ClusterStatusConditionDeleted)
}

func (s CommonClusterStatus) HasDeletingCondition() bool {
	return hasCondition(s.Conditions, ClusterStatusConditionDeleting)
}

func (s CommonClusterStatus) HasUpdatedCondition() bool {
	return hasCondition(s.Conditions, ClusterStatusConditionUpdated)
}

func (s CommonClusterStatus) HasUpdatingCondition() bool {
	return hasCondition(s.Conditions, ClusterStatusConditionUpdating)
}

func (s CommonClusterStatus) HasVersion(semver string) bool {
	return hasVersion(s.Versions, semver)
}

func (s CommonClusterStatus) LatestVersion() string {
	if len(s.Versions) == 0 {
		return ""
	}

	latest := s.Versions[0]

	for _, v := range s.Versions {
		if latest.LastTransitionTime.Time.Before(v.LastTransitionTime.Time) {
			latest = v
		}
	}

	return latest.Version
}

func (s CommonClusterStatus) WithCreatedCondition() []CommonClusterStatusCondition {
	return withCondition(s.Conditions, ClusterStatusConditionCreating, ClusterStatusConditionCreated, time.Now())
}

func (s CommonClusterStatus) WithCreatingCondition() []CommonClusterStatusCondition {
	return withCondition(s.Conditions, ClusterStatusConditionCreated, ClusterStatusConditionCreating, time.Now())
}

func (s CommonClusterStatus) WithDeletedCondition() []CommonClusterStatusCondition {
	return withCondition(s.Conditions, ClusterStatusConditionDeleting, ClusterStatusConditionDeleted, time.Now())
}

func (s CommonClusterStatus) WithDeletingCondition() []CommonClusterStatusCondition {
	return withCondition(s.Conditions, ClusterStatusConditionDeleted, ClusterStatusConditionDeleting, time.Now())
}

func (s CommonClusterStatus) WithNewVersion(version string) []CommonClusterStatusVersion {
	newVersion := CommonClusterStatusVersion{
		LastTransitionTime: DeepCopyTime{time.Now()},
		Version:            version,
	}

	return withVersion(s.Versions, newVersion, ClusterVersionLimit)
}

func (s CommonClusterStatus) WithUpdatedCondition() []CommonClusterStatusCondition {
	return withCondition(s.Conditions, ClusterStatusConditionUpdating, ClusterStatusConditionUpdated, time.Now())
}

func (s CommonClusterStatus) WithUpdatingCondition() []CommonClusterStatusCondition {
	return withCondition(s.Conditions, ClusterStatusConditionUpdated, ClusterStatusConditionUpdating, time.Now())
}

func getCondition(conditions []CommonClusterStatusCondition, condition string) CommonClusterStatusCondition {
	for _, c := range conditions {
		if c.Condition == condition {
			return c
		}
	}

	return CommonClusterStatusCondition{}
}

func hasCondition(conditions []CommonClusterStatusCondition, condition string) bool {
	for _, c := range conditions {
		if c.Condition == condition {
			return true
		}
	}

	return false
}

func hasVersion(versions []CommonClusterStatusVersion, search string) bool {
	for _, v := range versions {
		if v.Version == search {
			return true
		}
	}

	return false
}

func withCondition(conditions []CommonClusterStatusCondition, search string, replace string, t time.Time) []CommonClusterStatusCondition {
	newConditions := []CommonClusterStatusCondition{
		{
			LastTransitionTime: DeepCopyTime{t},
			Condition:          replace,
		},
	}

	for _, c := range conditions {
		if c.Condition == search {
			continue
		}

		newConditions = append(newConditions, c)
	}

	return newConditions
}

// withVersion computes a list of version history using the given list and new
// version structure to append. withVersion also limits total amount of elements
// in the list by cutting off the tail with respect to the limit parameter.
func withVersion(versions []CommonClusterStatusVersion, version CommonClusterStatusVersion, limit int) []CommonClusterStatusVersion {
	if hasVersion(versions, version.Version) {
		return versions
	}

	var newVersions []CommonClusterStatusVersion

	start := 0
	if len(versions) >= limit {
		start = len(versions) - limit + 1
	}

	sort.Sort(sortClusterStatusVersionsByDate(versions))

	for i := start; i < len(versions); i++ {
		newVersions = append(newVersions, versions[i])
	}

	newVersions = append(newVersions, version)

	return newVersions
}
