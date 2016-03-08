package semv

import (
	"strings"
	"testing"
)

// reversibleVersions are versions that when parsed, and
// String() is called on the resulting version, the original
// input string is returned.
var reversibleVersions = map[string]Version{
	"1":                          {1, 0, 0, "", "", Major},
	"1.2":                        {1, 2, 0, "", "", MajorMinor},
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

// invalidVersions list invalid versions and the expected error messages
var invalidVersions = map[string]string{
	"x.1.2": "unexpected character 'x' at position 0",
	"1.x.2": "unexpected character 'x' at position 2",
	"1.2.x": "unexpected character 'x' at position 4",
}

func TestString(t *testing.T) {
	for expectedString, inputVersion := range reversibleVersions {
		if s := inputVersion.String(); s != expectedString {
			t.Errorf("Got %+v.String() == %+v; expected %q", inputVersion, s, expectedString)
		}
	}
}

func TestParse(t *testing.T) {
	for inputString, expectedVersion := range reversibleVersions {
		actual, err := Parse(inputString)
		if err != nil {
			t.Error(err)
		}
		if actual != expectedVersion {
			t.Errorf("Got Parse(%q) == % +v; expected % +v", inputString, actual, expectedVersion)
		}
	}
}

var validParseAnyMap = map[string]string{
	"hellow world 1":         "1",
	"hellow world 1 1.2.3":   "1",
	"hi 1.2 world 9":         "1.2",
	"no 1-beta+meta":         "1-beta+meta",
	"yes 1.2.3-beta+meta!!!": "1.2.3-beta+meta",
	"yes 1.2+meta!!!":        "1.2+meta",
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

func TestParseError(t *testing.T) {
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
