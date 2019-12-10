package v1alpha2

import (
	"sort"
	"time"
)

func (s CommonClusterStatusCluster) GetCreatedCondition() CommonClusterStatusClusterCondition {
	return getCondition(s.Conditions, ClusterStatusConditionCreated)
}

func (s CommonClusterStatusCluster) GetCreatingCondition() CommonClusterStatusClusterCondition {
	return getCondition(s.Conditions, ClusterStatusConditionCreating)
}

func (s CommonClusterStatusCluster) GetDeletedCondition() CommonClusterStatusClusterCondition {
	return getCondition(s.Conditions, ClusterStatusConditionDeleted)
}

func (s CommonClusterStatusCluster) GetDeletingCondition() CommonClusterStatusClusterCondition {
	return getCondition(s.Conditions, ClusterStatusConditionDeleting)
}

func (s CommonClusterStatusCluster) GetUpdatedCondition() CommonClusterStatusClusterCondition {
	return getCondition(s.Conditions, ClusterStatusConditionUpdated)
}

func (s CommonClusterStatusCluster) GetUpdatingCondition() CommonClusterStatusClusterCondition {
	return getCondition(s.Conditions, ClusterStatusConditionUpdating)
}

func (s CommonClusterStatusCluster) HasCreatedCondition() bool {
	return hasCondition(s.Conditions, ClusterStatusConditionCreated)
}

func (s CommonClusterStatusCluster) HasCreatingCondition() bool {
	return hasCondition(s.Conditions, ClusterStatusConditionCreating)
}

func (s CommonClusterStatusCluster) HasDeletedCondition() bool {
	return hasCondition(s.Conditions, ClusterStatusConditionDeleted)
}

func (s CommonClusterStatusCluster) HasDeletingCondition() bool {
	return hasCondition(s.Conditions, ClusterStatusConditionDeleting)
}

func (s CommonClusterStatusCluster) HasUpdatedCondition() bool {
	return hasCondition(s.Conditions, ClusterStatusConditionUpdated)
}

func (s CommonClusterStatusCluster) HasUpdatingCondition() bool {
	return hasCondition(s.Conditions, ClusterStatusConditionUpdating)
}

func (s CommonClusterStatusCluster) HasVersion(semver string) bool {
	return hasVersion(s.Versions, semver)
}

func (s CommonClusterStatusCluster) LatestCondition() string {
	if len(s.Conditions) == 0 {
		return ""
	}

	sort.Sort(sort.Reverse(sortClusterStatusConditionsByDate(s.Conditions)))

	return s.Conditions[0].Condition
}

func (s CommonClusterStatusCluster) LatestVersion() string {
	if len(s.Versions) == 0 {
		return ""
	}

	sort.Sort(sort.Reverse(sortClusterStatusVersionsByDate(s.Versions)))

	return s.Versions[0].Version
}

func (s CommonClusterStatusCluster) WithCreatedCondition() []CommonClusterStatusClusterCondition {
	newCondition := CommonClusterStatusClusterCondition{
		LastTransitionTime: DeepCopyTime{time.Now()},
		Condition:          ClusterStatusConditionCreated,
	}

	return withCondition(s.Conditions, newCondition, ClusterConditionLimit)
}

func (s CommonClusterStatusCluster) WithCreatingCondition() []CommonClusterStatusClusterCondition {
	newCondition := CommonClusterStatusClusterCondition{
		LastTransitionTime: DeepCopyTime{time.Now()},
		Condition:          ClusterStatusConditionCreating,
	}

	return withCondition(s.Conditions, newCondition, ClusterConditionLimit)
}

func (s CommonClusterStatusCluster) WithDeletedCondition() []CommonClusterStatusClusterCondition {
	newCondition := CommonClusterStatusClusterCondition{
		LastTransitionTime: DeepCopyTime{time.Now()},
		Condition:          ClusterStatusConditionDeleted,
	}

	return withCondition(s.Conditions, newCondition, ClusterConditionLimit)
}

func (s CommonClusterStatusCluster) WithDeletingCondition() []CommonClusterStatusClusterCondition {
	newCondition := CommonClusterStatusClusterCondition{
		LastTransitionTime: DeepCopyTime{time.Now()},
		Condition:          ClusterStatusConditionDeleting,
	}

	return withCondition(s.Conditions, newCondition, ClusterConditionLimit)
}

func (s CommonClusterStatusCluster) WithNewVersion(version string) []CommonClusterStatusClusterVersion {
	newVersion := CommonClusterStatusClusterVersion{
		LastTransitionTime: DeepCopyTime{time.Now()},
		Version:            version,
	}

	return withVersion(s.Versions, newVersion, ClusterVersionLimit)
}

func (s CommonClusterStatusCluster) WithUpdatedCondition() []CommonClusterStatusClusterCondition {
	newCondition := CommonClusterStatusClusterCondition{
		LastTransitionTime: DeepCopyTime{time.Now()},
		Condition:          ClusterStatusConditionUpdated,
	}

	return withCondition(s.Conditions, newCondition, ClusterConditionLimit)
}

func (s CommonClusterStatusCluster) WithUpdatingCondition() []CommonClusterStatusClusterCondition {
	newCondition := CommonClusterStatusClusterCondition{
		LastTransitionTime: DeepCopyTime{time.Now()},
		Condition:          ClusterStatusConditionUpdating,
	}

	return withCondition(s.Conditions, newCondition, ClusterConditionLimit)
}

func getCondition(conditions []CommonClusterStatusClusterCondition, condition string) CommonClusterStatusClusterCondition {
	for _, c := range conditions {
		if c.Condition == condition {
			return c
		}
	}

	return CommonClusterStatusClusterCondition{}
}

func getConditionForPair(a CommonClusterStatusClusterCondition) string {
	for _, p := range conditionPairs {
		if p[0] == a.Condition {
			return p[1]
		}
		if p[1] == a.Condition {
			return p[0]
		}
	}

	return ""
}

func hasCondition(conditions []CommonClusterStatusClusterCondition, condition string) bool {
	for _, c := range conditions {
		if c.Condition == condition {
			return true
		}
	}

	return false
}

func hasVersion(versions []CommonClusterStatusClusterVersion, search string) bool {
	for _, v := range versions {
		if v.Version == search {
			return true
		}
	}

	return false
}

func isConditionPair(a CommonClusterStatusClusterCondition, b CommonClusterStatusClusterCondition) bool {
	for _, p := range conditionPairs {
		if p[0] == a.Condition && p[1] == b.Condition {
			return true
		}
		if p[1] == a.Condition && p[0] == b.Condition {
			return true
		}
	}

	return false
}

// withCondition takes a list of status conditions and manages the given list
// according to the condition to add on top and the given limit argument. The
// limit argument should always only be given by ClusterConditionLimit. Also see
// the godoc there. The limit is applied to condition pairs as defined by
// conditionPairs. Internally the given conditions list is copied so that the
// input arguments are not manipulated by accident. One specific functionality
// of withCondition is that incomplete condition pairs are completed
// automatically as this may happen due to unexpected behaviour in the callers
// environment. For more information on implementation details read the inline
// comments of the code.
func withCondition(conditions []CommonClusterStatusClusterCondition, condition CommonClusterStatusClusterCondition, limit int) []CommonClusterStatusClusterCondition {
	// We create a new list which acts like a copy so the input parameters are not
	// manipulated. Here we also prepend the given condition and inject certain
	// missing conditions in case the condition list gets out of sync
	// unintendedly due to any eventual bugs. Test case 8 demonstrates that.
	var newConditions []CommonClusterStatusClusterCondition
	{
		if len(conditions) > 0 && conditions[0].Condition == condition.Condition {
			injected := CommonClusterStatusClusterCondition{
				// The implication of unintendedly untracked conditions is that the
				// automatically added condition does not obtain a reasonable timestamp.
				// Here we take the timestamp of the new condition we want to track and
				// substract one nano second from it to keep the order intact.
				LastTransitionTime: DeepCopyTime{condition.LastTransitionTime.Add(-(1 * time.Nanosecond))},
				Condition:          getConditionForPair(condition),
			}
			newConditions = append(newConditions, injected)
		}

		newConditions = append(newConditions, condition)

		for _, c := range conditions {
			newConditions = append(newConditions, c)
		}
	}

	// The new list is sorted to have the first item being the oldest. This is to
	// have an easier grouping mechanism below. When the first item of a new pair
	// is added, it would throw of the grouping when the order would be kept as
	// given.
	sort.Sort(sortClusterStatusConditionsByDate(newConditions))

	// The conditions are grouped into their corresponding pairs of transitioning
	// states. Associated Creating/Created, Updating/Updated and Deleting/Deleted
	// conditions are put together.
	var conditionGroups [][]CommonClusterStatusClusterCondition
	for len(newConditions) > 0 {
		var g []CommonClusterStatusClusterCondition

		for _, c := range newConditions {
			// If the list only contains one item anymore, we process it separately
			// here and be done. Otherwhise the pruning of the list below panics due
			// to the range calculations.
			if len(newConditions) == 1 {
				g = append(g, c)
				newConditions = []CommonClusterStatusClusterCondition{}
				break
			}

			// Put the first item from the top of the list into the group and drop
			// the grouped item from the list.
			if len(g) == 0 {
				g = append(g, c)
				newConditions = newConditions[1:len(newConditions)]
				continue
			}

			// When we find the second item of the pair we are done for this group.
			if len(g) == 1 {
				if isConditionPair(g[0], c) {
					g = append(g, c)
					newConditions = newConditions[1:len(newConditions)]
				}
				break
			}
		}

		conditionGroups = append(conditionGroups, g)
	}

	// The pairs are now grouped. When there are only three group kinds for
	// create/update/delete, conditionPairs has a length of 3. Each of the groups
	// has then as many pairs as grouped together. Below these groups are limited.
	var conditionPairs [][]CommonClusterStatusClusterCondition
	for len(conditionGroups) > 0 {
		var p []CommonClusterStatusClusterCondition

		for _, g := range conditionGroups {
			if len(p) == 0 {
				p = append(p, g...)
				conditionGroups = conditionGroups[1:len(conditionGroups)]
				continue
			}

			if len(g) >= 1 {
				if isConditionPair(p[0], g[0]) || isConditionPair(p[1], g[0]) {
					p = append(p, g...)
					conditionGroups = conditionGroups[1:len(conditionGroups)]
				}
			}
		}

		conditionPairs = append(conditionPairs, p)
	}

	// Here the list is finally flattened again and its pairs are limitted to the
	// input parameter.
	var limittedList []CommonClusterStatusClusterCondition
	for _, p := range conditionPairs {
		// We compute the pair limit here for the total number of items. This is why
		// we multiply by 2. When the limit is 5, we want to track for instance 5
		// Updating/Updated pairs. Additionally when there is an item of a new pair
		// and the list must be capped, the additional odd of the new item has to be
		// considered when computing the limit. This results in an additional pair
		// being dropped. Test case 6 demonstrates that.
		l := (limit * 2) - (len(p) % 2)
		if len(p) < l {
			l = len(p)
		}

		limittedList = append(limittedList, p[len(p)-l:len(p)]...)
	}

	// We reverse the list order to have the item with the highest timestamp at
	// the top again.
	sort.Sort(sort.Reverse(sortClusterStatusConditionsByDate(limittedList)))

	return limittedList
}

// withVersion computes a list of version history using the given list and new
// version structure to append. withVersion also limits the total amount of
// elements in the list by cutting off the tail with respect to the limit
// parameter.
func withVersion(versions []CommonClusterStatusClusterVersion, version CommonClusterStatusClusterVersion, limit int) []CommonClusterStatusClusterVersion {
	if hasVersion(versions, version.Version) {
		return versions
	}

	// Create a copy to not manipulate the input list.
	var newVersions []CommonClusterStatusClusterVersion
	for _, v := range versions {
		newVersions = append(newVersions, v)
	}

	// Sort the versions in a way that the newest version, namely the one with the
	// highest timestamp, is at the top of the list.
	sort.Sort(sort.Reverse(sortClusterStatusVersionsByDate(newVersions)))

	// Calculate the index for capping the list in the next step.
	l := limit - 1
	if len(newVersions) < l {
		l = len(newVersions)
	}

	// Cap the list and prepend the new version.
	newVersions = append([]CommonClusterStatusClusterVersion{version}, newVersions[0:l]...)

	return newVersions
}
