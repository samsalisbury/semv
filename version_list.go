package semv

import "sort"

// VersionList is a slice of Versions, with some extra functions...
type VersionList []Version

// Len returns the number of versions in this list.
func (vl VersionList) Len() int { return len(vl) }

// Swap is used by sort to swap elements in the list.
func (vl VersionList) Swap(i, j int) { vl[i], vl[j] = vl[j], vl[i] }

// Less indicates whether or not the Version at index i is less than that at
// index j.
func (vl VersionList) Less(i, j int) bool { return vl[i].Less(vl[j]) }

// Sorted returns a VersionList sorted from lowest to highest versions.
func (vl VersionList) Sorted() VersionList {
	sort.Sort(vl)
	return vl
}

// SortedDesc returns a VersionList sorted in the opposite direction to
// that returned from Sort.
func (vl VersionList) SortedDesc() VersionList {
	sort.Sort(vl)
	sort.Reverse(vl)
	return vl
}

// GreatestSatisfying returns the greatest (highest) version contained in the
// VersionList, which satisfies the passed Range. If none are found that satisfy
// the range, the second return value is false, otherwise it is true.
func (vl VersionList) GreatestSatisfying(r Range) (Version, bool) {
	for _, v := range vl.SortedDesc() {
		if r.SatisfiedBy(v) {
			return v, true
		}
	}
	return Version{}, false
}
