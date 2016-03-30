package semv

import (
	"encoding/json"
	"strings"
	"testing"
)

// reversibleParseVersions are versions that when parsed, and
// String() is called on the resulting version, the original
// input string is returned.
var reversibleParseVersions = map[string]Version{
	"1":                          {1, 0, 0, "", "", Major},
	"1.2":                        {1, 2, 0, "", "", MajorMinor},
	"1.2.3":                      {1, 2, 3, "", "", MajorMinorPatch},
	"1.2.3-beta.1":               {1, 2, 3, "beta.1", "", MMPPre},
	"1.2.3-beta.1+some.metadata": {1, 2, 3, "beta.1", "some.metadata", Complete},
	"0.0.0":                                              {0, 0, 0, "", "", MajorMinorPatch},
	"0.0.0-beta":                                         {0, 0, 0, "beta", "", MMPPre},
	"0.0.100-beta.1":                                     {0, 0, 100, "beta.1", "", MMPPre},
	"0.100.100-beta.1+some.metadata":                     {0, 100, 100, "beta.1", "some.metadata", Complete},
	"100.100.100-beta.1+some.metadata":                   {100, 100, 100, "beta.1", "some.metadata", Complete},
	"100.100.100-beta-dash-21+some.metadata":             {100, 100, 100, "beta-dash-21", "some.metadata", Complete},
	"100.100.100-beta-dash-21+some-dashing--metadata.45": {100, 100, 100, "beta-dash-21", "some-dashing--metadata.45", Complete},
}

func TestString(t *testing.T) {
	for expectedString, inputVersion := range reversibleParseVersions {
		if s := inputVersion.String(); s != expectedString {
			t.Errorf("Got %+v.String() == %+v; expected %q", inputVersion, s, expectedString)
		}
	}
}

func TestParse(t *testing.T) {
	for inputString, expectedVersion := range reversibleParseVersions {
		actual, err := Parse(inputString)
		if err != nil {
			t.Error(err)
		}
		if actual != expectedVersion {
			t.Errorf("Got Parse(%q) == % +v; expected % +v", inputString, actual, expectedVersion)
		}
	}
}

// reversibleParseExactVersions are versions that successfully parse with
// ParseExactSemver2_0_0, and when
// String() is called on the resulting version, the original
// input string is returned.
var parseExactVersions = map[string]Version{
	"1.2.3":                      {1, 2, 3, "", "", MajorMinorPatch},
	"1.2.3-beta.1":               {1, 2, 3, "beta.1", "", MMPPre},
	"1.2.3-beta.1+some.metadata": {1, 2, 3, "beta.1", "some.metadata", Complete},
	"0.0.0":                                              {0, 0, 0, "", "", MajorMinorPatch},
	"0.0.100-beta.1":                                     {0, 0, 100, "beta.1", "", MMPPre},
	"0.100.100-beta.1+some.metadata":                     {0, 100, 100, "beta.1", "some.metadata", Complete},
	"100.100.100-beta.1+some.metadata":                   {100, 100, 100, "beta.1", "some.metadata", Complete},
	"100.100.100-beta-dash-21+some.metadata":             {100, 100, 100, "beta-dash-21", "some.metadata", Complete},
	"100.100.100-beta-dash-21+some-dashing--metadata.45": {100, 100, 100, "beta-dash-21", "some-dashing--metadata.45", Complete},
}

func TestParseExactSemver2_0_0(t *testing.T) {
	for inputString, expectedVersion := range parseExactVersions {
		actual, err := ParseExactSemver2(inputString)
		if err != nil {
			t.Error(err)
		}
		if actual != expectedVersion {
			t.Errorf("Got Parse(%q) == % +v; expected % +v", inputString, actual, expectedVersion)
		}
	}
}

var validParseAnyMap = map[string]string{
	"go version go1.6 darwin/amd64":   "1.6",
	"go version go1.5.3 darwin/amd64": "1.5.3",
	"git version 2.6.1":               "2.6.1",
	"hello world 1":                   "1",
	"hello world 1 1.2.3":             "1",
	"hi 1.2 world 9":                  "1.2",
	"no 1-beta+meta":                  "1-beta+meta",
	"yes 1.2.3-beta+meta!!!":          "1.2.3-beta+meta",
	"yes 1.2+meta!!!":                 "1.2+meta",
}

func TestParseAny(t *testing.T) {
	for input, expected := range validParseAnyMap {
		actual, err := ParseAny(input)
		if err != nil {
			t.Errorf("unexpected error %s", err)
			continue
		}
		if actual.String() != expected {
			t.Errorf("got version %q from %q; expected %q", actual, input, expected)
		}
	}
}

// invalidVersions list invalid versions and the expected error messages
var invalidVersions = map[string]string{
	"x.1.2": "unexpected character 'x' at position 0",
	"1.x.2": "unexpected character 'x' at position 2",
	"1.2.x": "unexpected character 'x' at position 4",
}

func TestParseErrors(t *testing.T) {
	for inputString, expectedError := range invalidVersions {
		_, err := Parse(inputString)
		if err == nil {
			t.Errorf("successfully parsed invalid string %q as version", inputString)
			continue
		}
		if !strings.Contains(err.Error(), expectedError) {
			t.Errorf("got error message %q; expected %q", err.Error(), expectedError)
		}
	}
}

var invalidExactSemver2_0_0Versions = map[string]string{
	"01.2.3":    "unexpected preceding zero in major component",
	"1.02.3":    "unexpected preceding zero in minor component",
	"1.2.03":    "unexpected preceding zero in patch component",
	"00.0.0":    "unexpected preceding zero in major component",
	"1.2":       "version incomplete: missing patch component",
	"1":         "version incomplete: missing minor component",
	"x":         "unexpected character 'x' at position 0",
	"1.2.y":     "unexpected character 'y' at position 4",
	".2.3":      "zero-length major component",
	"1..3":      "zero-length minor component",
	"1.2.-beta": "zero-length patch component",
}

func TestParseExactSemver2_0_0Error(t *testing.T) {
	for input, expectedError := range invalidExactSemver2_0_0Versions {
		_, err := ParseExactSemver2(input)
		if err == nil {
			t.Errorf("successfully parsed invalid semver 2.0.0 string %q as version", input)
		}
		if !strings.Contains(err.Error(), expectedError) {
			t.Errorf("got error message %q; expected %q", err.Error(), expectedError)
		}
	}
}

// orderedVersions uses the specific example from §11 of the semver spec at
// http://semver.org/spec/v2.0.0.html
var orderedVersions = []Version{
	MustParseExactSemver2("1.0.0-alpha"),
	MustParseExactSemver2("1.0.0-alpha.1"),
	MustParseExactSemver2("1.0.0-alpha.beta"),
	MustParseExactSemver2("1.0.0-beta"),
	MustParseExactSemver2("1.0.0-beta.2"),
	MustParseExactSemver2("1.0.0-beta.11"),
	MustParseExactSemver2("1.0.0-rc.1"),
	MustParseExactSemver2("1.0.0"),
}

func TestLess(t *testing.T) {
	for i := 1; i < len(orderedVersions); i++ {
		lesser := orderedVersions[i-1]
		greater := orderedVersions[i]
		if !lesser.Less(greater) {
			t.Errorf("expected %q to be less than %q", lesser, greater)
		}
		if greater.Less(lesser) {
			t.Errorf("expected %q not to be less than %q", greater, lesser)
		}
	}
}

func TestLess2(t *testing.T) {
	if !MustParse("1.2.3-beta").Less(MustParse("1.2.3")) {
		t.Errorf("expected 1.0.0-beta to be less than 1.0.0")
	}
}

var equalVersions = []Version{
	MustParseExactSemver2("1.0.0-alpha.1.4-beta.2+abc"),
	MustParseExactSemver2("1.0.0-alpha.1.4-beta.2+123"),
	MustParseExactSemver2("1.0.0-alpha.1.4-beta.2+gsfdnjfdhisg9efwd897ywrfwerf"),
	MustParseExactSemver2("1.0.0-alpha.1.4-beta.2+abc.def.ghy.123"),
	MustParseExactSemver2("1.0.0-alpha.1.4-beta.2+123.456.789.abc"),
}

func TestEquals(t *testing.T) {
	for _, v := range equalVersions {
		for _, to := range equalVersions {
			if !v.Equals(to) {
				t.Errorf("expected %q to equal %q", v, to)
			}
		}
	}
}

func TestNotEquals(t *testing.T) {
	if MustParse("0.0.0-beta") == MustParse("0.0.0") {
		t.Errorf("expected 0.0.0-beta ≠ 0.0.0")
	}
}

func TestIncrements(t *testing.T) {
	v1 := MustParse("1.0.0")
	v1_0_1 := v1.IncrementPatch()
	if v1_0_1.String() != "1.0.1" {
		t.Errorf("expected increment to 1.0.1; got %q", v1)
	}
	v1_1_0 := v1_0_1.IncrementMinor()
	if v1_1_0.String() != "1.1.0" {
		t.Errorf("expected increment minor to 1.1.0; got %q", v1)
	}
	v1_1_1 := v1_1_0.IncrementPatch()
	if v1_1_1.String() != "1.1.1" {
		t.Errorf("expected increment patch to 1.1.1; got %q", v1)
	}
	v2_0_0 := v1_1_1.IncrementMajor()
	if v2_0_0.String() != "2.0.0" {
		t.Errorf("expected increment major to 2.0.0; got %q", v1)
	}
}

func (v Version) dump() string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}
