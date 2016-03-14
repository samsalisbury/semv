package semv

import (
	"encoding/json"
	"testing"
)

var (
	v1_0_0 = MustParse("1.0.0")
	v2_0_0 = MustParse("2.0.0")

	validRanges = map[string]Range{
		"1.0.0":    EqualTo(v1_0_0),
		"=1.0.0":   EqualTo(v1_0_0),
		"==1.0.0":  EqualTo(v1_0_0),
		"== 1":     EqualTo(v1_0_0),
		"<1.0.0":   LessThan(v1_0_0),
		"> 1.0.0":  GreaterThan(v1_0_0),
		">= 2.0.0": GreaterThanOrEqualTo(v2_0_0),
		"<=2.0.0":  LessThanOrEqualTo(v2_0_0),
		"^1.0.0":   GreaterThanOrEqualToAndLessThan(v1_0_0, v2_0_0),
		"~1.0.0":   GreaterThanOrEqualToAndLessThan(v1_0_0, v1_0_0.IncrementMinor()),
	}
)

func TestEqualityToSelf(t *testing.T) {
	ranges := []Range{
		EqualTo(v1_0_0),
		LessThan(v1_0_0),
		GreaterThan(v1_0_0),
		GreaterThanOrEqualTo(v1_0_0),
		LessThanOrEqualTo(v1_0_0),
		GreaterThanOrEqualToAndLessThan(v1_0_0, v2_0_0),
	}

	for _, r := range ranges {
		if !r.Equals(r) {
			t.Errorf("expected %q to equal %q", r, r)
		}
	}
}

func TestEqualityToEqualCopy(t *testing.T) {
	rangeFuncs := []func() Range{
		func() Range { return EqualTo(v1_0_0) },
		func() Range { return LessThan(v1_0_0) },
		func() Range { return GreaterThan(v1_0_0) },
		func() Range { return GreaterThanOrEqualTo(v1_0_0) },
		func() Range { return LessThanOrEqualTo(v1_0_0) },
		func() Range { return GreaterThanOrEqualToAndLessThan(v1_0_0, v2_0_0) },
	}

	for _, f := range rangeFuncs {
		r1 := f()
		r2 := f()
		if !r1.Equals(r2) || !r2.Equals(r1) {
			t.Errorf("expected %q to equal %q", r1, r2)
		}
	}
}

func TestParseRange_Valid(t *testing.T) {
	for inputString, expected := range validRanges {
		actual, err := ParseRange(inputString)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if !actual.Equals(expected) {
			t.Errorf("parse range %q gave %q; wanted %q", inputString, actual, expected)
		}
	}
}

func TestParseRange_Invalid(t *testing.T) {
	r, err := ParseRange("")
	if err == nil {
		t.Errorf(`ParseRange("") did not return an error`)
	}
	if (r != Range{}) {
		t.Errorf(`ParseRange("") did not return a zeroed range`)
	}
}

var rangesToStrings = map[Range]string{
	LessThan(v1_0_0):                                                 "<1.0.0",
	GreaterThan(v1_0_0):                                              ">1.0.0",
	GreaterThanOrEqualTo(v1_0_0):                                     ">=1.0.0",
	LessThanOrEqualTo(v1_0_0):                                        "<=1.0.0",
	GreaterThanOrEqualToAndLessThan(v1_0_0, v2_0_0):                  "^1.0.0",
	GreaterThanOrEqualToAndLessThan(v1_0_0, v1_0_0.IncrementMinor()): "~1.0.0",
}

func TestRangeString(t *testing.T) {
	for inputRange, expectedString := range rangesToStrings {
		actual := inputRange.String()
		if actual != expectedString {
			t.Errorf("got range string %q; expected %q", actual, expectedString)
		}
	}
}

var rangesToSatisfactoryVersions = map[string][]string{
	"1.0.0":          {"1.0.0", "1.0.0+whatever"},
	"^1.0.0":         {"1.99.99", "1.999.0+whatever"},
	"~1.0.0":         {"1.0.9999", "1.0.9999+whatever"},
	">1.0.0":         {"1.0.1", "2.0.0", "10009.0.0", "1.1.0+whatever"},
	"<1.0.0":         {"0.9.9", "0.0.1", "0.0.0"},
	">=1.0.0":        {"1.0.0", "1.0.1", "1.1.0", "2.0.0", "999.0.0"},
	"<=1.0.0":        {"1.0.0", "0.0.9", "0.999.999", "0.0.0", "0.0.1"},
	">0.2.3-beta.2":  {"0.2.3-beta.3", "0.2.3-beta.4", "0.2.3-rc", "0.2.3-rc.2", "0.2.4", "0.3.0"},
	">=0.2.3-beta.2": {"0.2.3-beta.2", "0.2.3-beta.3", "0.2.3-beta.4", "0.2.3-rc", "0.2.3-rc.2", "0.2.4", "0.3.0"},
	"~4.3.2-beta.3":  {"4.3.2-beta.3", "4.3.2-beta.4", "4.3.2-rc", "4.3.2-rc.1", "4.3.4", "4.3.9"},
	"^4.3.2-beta.3":  {"4.3.2-beta.3", "4.3.2-beta.4", "4.3.2-rc", "4.3.2-rc.1", "4.3.4", "4.3.9", "4.4.0", "4.9.9"},
}

func TestIsSatisfiedBy(t *testing.T) {
	for rangeString, versionStrings := range rangesToSatisfactoryVersions {
		r, err := ParseRange(rangeString)
		if err != nil {
			t.Errorf("unexpected range parsing error: %s", err)
			continue
		}
		for _, vs := range versionStrings {
			v, err := Parse(vs)
			if err != nil {
				t.Errorf("unexpected version parsing error: %s", err)
				continue
			}
			if !r.SatisfiedBy(v) {
				t.Errorf("expected range %q to be satisfied by version %q", r, v)
			}
		}
	}
}

var rangesToUnsatisfactoryVersions = map[string][]string{
	"1.0.0":          {"1.0.0-beta", "1.0.1"},
	"^1.2.3":         {"1.2.3-beta", "2.0.0", "1.0.0", "0.0.0"},
	"~1.2.3":         {"1.2.3-beta", "1.2.2", "1.3.0"},
	">1.0.0":         {"1.0.0", "1.0.0-beta", "0.0.0", "0.9.9", "0.0.1"},
	"<1.0.0":         {"1.0.0", "1.0.1", "2.1.0"},
	">=1.0.0":        {"1.0.0-beta", "0.9.9", "0.0.1", "0.0.0"},
	"<=1.0.0":        {"1.0.1", "1.1.0", "1.0.1-beta"},
	">0.2.3-beta.2":  {"0.2.3-beta.2", "0.2.2-beta.4", "0.2.3-beta"},
	">=0.2.3-beta.2": {"0.2.3-beta.1", "0.2.3-beta.0", "0.2.3-beta", "0.2.2-rc", "0.2.1"},
	"~4.3.2-beta.3":  {"4.3.2-alpha.2", "4.3.2-beta.2", "4.4.0-rc.1", "4.4.0", "4.5.1"},
	"^4.3.2-beta.3":  {"4.3.1-beta.3", "4.3.2-alpha", "5.0.0-alpha", "5.0.0", "5.0.1-alpha", "5.0.1"},
}

func TestIsNotSatisfiedBy(t *testing.T) {
	for rangeString, versionStrings := range rangesToUnsatisfactoryVersions {
		r, err := ParseRange(rangeString)
		if err != nil {
			t.Errorf("unexpected range parsing error: %s", err)
			continue
		}
		for _, vs := range versionStrings {
			v, err := Parse(vs)
			if err != nil {
				t.Errorf("unexpected version parsing error: %s", err)
				continue
			}
			if r.SatisfiedBy(v) {
				t.Errorf("expected range %q not to be satisfied by version %q:\nRange: %s\nVersion:%s",
					r, v, r.dump(), v.dump())
			}
		}
	}
}

func (r Range) dump() string {
	b, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return string(b)
}
