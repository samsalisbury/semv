package semv

import (
	"math/rand"
	"sort"
	"testing"
	"time"
)

func newOrderedVersionList() VersionList {
	vl := MustParseList(
		"0.0.0", "0.0.1", "0.0.2-beta", "0.0.2-beta.9", "0.0.2-beta.11", "0.0.2-rc.1", "0.0.2-rc.2", "0.0.2", "0.0.3-alpha", "0.0.3", "0.1.2",
		"0.1.3-beta", "0.1.4", "0.1.10", "0.1.11-beta", "0.1.11", "0.1.12-beta", "0.2.5", "0.2.35", "0.3.1", "1.0.0-beta", "1.0.0-beta.2",
		"1.0.0", "1.0.1", "1.0.2-beta.66", "1.0.2", "1.1.0-alpha.1", "1.1.0-alpha.2", "1.1.0-beta", "1.1.0-beta.2", "1.1.0", "1.1.1", "1.1.2",
		"1.1.3", "1.1.9", "1.2.0-beta", "1.2.0", "1.2.1", "2.0.0", "2.0.1", "2.1.1", "3.0.0", "3.5.6", "5.8.0")
	if !sort.IsSorted(vl) {
		panic("manually sorted version list is not correctly sorted")
	}
	return vl
}

var rangeToGreatestSatisfyingVersion = map[string]string{
	"0.0.0":         "0.0.0",
	"0.0.2-beta":    "0.0.2-beta",
	"<1.1.9":        "1.1.3",
	"<1.2.0":        "1.1.9",
	"<0.0.1":        "0.0.0",
	">1.2.0":        "5.8.0",
	">0.0.0":        "5.8.0",
	"<=1.1.9":       "1.1.9",
	"<=0.0.0":       "0.0.0",
	"<=0.0.1":       "0.0.1",
	">=0.1.11-beta": "5.8.0",
	">=1.1.3":       "5.8.0",
	">=1.2.1":       "5.8.0",
	">=5.8.0":       "5.8.0",
	"~0.0.0":        "0.0.3",
	"~0.0.3":        "0.0.3",
	"~0.1.5":        "0.1.11",
	"~0.0.1":        "0.0.3",
	"~1.1.0-beta":   "1.1.9",
	"~1.2.0":        "1.2.1",
	"~1.2.1":        "1.2.1",
	"^0.0.0":        "0.3.1",
	"^0.0.3":        "0.3.1",
	"^0.1.11-beta":  "0.3.1",
	"^1.0.0":        "1.2.1",
	"^1.1.2-rc.2":   "1.2.1",
	"^1.1.9":        "1.2.1",
	"^1.2.0":        "1.2.1",
	"^2.0.0":        "2.1.1",
	"^2.1.1":        "2.1.1",
	"^3.0.0":        "3.5.6",
}

func newRandomisedVersionList() VersionList {
	vl := newOrderedVersionList()
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 999; i++ {
		for i := range vl {
			j := rand.Intn(i + 1)
			vl.Swap(i, j)
		}
		if !sort.IsSorted(vl) {
			return vl
		}
	}
	panic("failed to randomise version list in 999 attempts")
}

func TestSortedDesc(t *testing.T) {
	randomised := newRandomisedVersionList()
	sortedAsc := randomised.Sorted()
	sortedDesc := randomised.SortedDesc()

	failed := false
	for i, ascV := range sortedAsc {
		j := len(sortedDesc) - (i + 1)
		descV := sortedDesc[j]
		t.Logf("%q\t%q", sortedAsc[i], sortedDesc[i])
		if !ascV.Equals(descV) {
			failed = true
		}
	}
	if failed {
		t.Error("one or more items was out of expected order, see logs above")
	}
}

func TestSorted(t *testing.T) {

	ordered := newOrderedVersionList()
	randomised := newRandomisedVersionList()

	sort.Sort(randomised)
	sorted := randomised

	for i, expected := range ordered {
		actual := sorted[i]
		if !actual.Equals(expected) {
			t.Errorf("sort order incorrect, got %q at position %d; want %q", actual, i, expected)
		}
	}
}

func TestGreatestSatisfying(t *testing.T) {
	// run these tests in deterministic order:
	orderedKeys := make([]string, len(rangeToGreatestSatisfyingVersion))
	i := 0
	for key := range rangeToGreatestSatisfyingVersion {
		orderedKeys[i] = key
		i++
	}
	sort.Strings(orderedKeys)
	// actual test
	for _, rangeString := range orderedKeys {
		versionString := rangeToGreatestSatisfyingVersion[rangeString]
		r := MustParseRange(rangeString)
		expected := MustParse(versionString)
		vl := newRandomisedVersionList()
		actual, ok := vl.GreatestSatisfying(r)
		if !ok {
			t.Errorf("expected to find a version satisfying %q", r)
			continue
		}
		if actual != expected {
			t.Errorf("got greatest version %q satisfying %q; expected %q", actual, r, expected)
		}
	}
}
