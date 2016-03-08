package semv

import "testing"

var randomised = VersionList{
	MustParse("0.9.2"),
	MustParse("1.0.0"),
	MustParse("1.1.1"),
	MustParse("1.0.0-beta"),
	MustParse("0.9.9-alpha.3"),
	MustParse("0.9.8-rc.1"),
	MustParse("0.9.9-alpha.12"),
	MustParse("0.9.0"),
	MustParse("0.9.3"),
	MustParse("0.9.9-alpha.9"),
	MustParse("0.9.9-beta.1"),
	MustParse("0.9.9-beta"),
	MustParse("0.9.1"),
	MustParse("1.1.1-beta.2"),
}

var expectedOrder = VersionList{
	MustParse("0.9.0"),
	MustParse("0.9.1"),
	MustParse("0.9.2"),
	MustParse("0.9.3"),
	MustParse("0.9.8-rc.1"),
	MustParse("0.9.9-alpha.3"),
	MustParse("0.9.9-alpha.9"),
	MustParse("0.9.9-alpha.12"),
	MustParse("0.9.9-beta"),
	MustParse("0.9.9-beta.1"),
	MustParse("1.0.0-beta"),
	MustParse("1.0.0"),
	MustParse("1.1.1-beta.2"),
	MustParse("1.1.1"),
}

func TestSorted(t *testing.T) {
	sorted := randomised.Sorted()
	for i, actual := range sorted {
		expected := expectedOrder[i]
		if !actual.Equals(expected) {
			t.Errorf("sort order incorrect, got %q at position %d; want %q",
				actual, i, expected)
		}
	}
}

func TestGreatestSatisfying(t *testing.T) {

}
