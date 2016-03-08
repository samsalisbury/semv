package semv

import "testing"

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

func TestParseRange(t *testing.T) {
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
