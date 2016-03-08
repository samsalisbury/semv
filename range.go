package semv

import "fmt"

type (
	// Range is a semver range.
	Range struct {
		Min, MinEqual, Max, MaxEqual *Version
	}
)

func ParseRange(s string) (Range, error) {
	v, err := ParseAny(s)
	if err != nil {
		return Range{}, err
	}
	switch s[:2] {
	case "==":
		return EqualTo(v), nil
	case ">=":
		return GreaterThanOrEqualTo(v), nil
	case "<=":
		return LessThanOrEqualTo(v), nil
	}
	switch s[0] {
	case '=', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return EqualTo(v), nil
	case '>':
		return GreaterThan(v), nil
	case '<':
		return LessThan(v), nil
	case '~':
		max := v
		max.Minor++
		max.Patch = 0
		return Range{
			MinEqual: &v,
			Max:      &max,
		}, nil
	case '^':
		max := v
		max.Major++
		max.Minor = 0
		max.Patch = 0
		return Range{
			MinEqual: &v,
			Max:      &max,
		}, nil
	}
	return Range{}, fmt.Errorf("unable to parse version range %q", s)
}

func GreaterThan(v Version) Range {
	return Range{Min: &v}
}

func LessThan(v Version) Range {
	return Range{Max: &v}
}

func EqualTo(v Version) Range {
	return Range{MinEqual: &v, MaxEqual: &v}
}

func GreaterThanOrEqualTo(v Version) Range {
	return Range{MinEqual: &v}
}

func LessThanOrEqualTo(v Version) Range {
	return Range{MaxEqual: &v}
}

func GreaterThanOrEqualToAndLessThan(min, lessThan Version) Range {
	return Range{MinEqual: &min, Max: &lessThan}
}

func (r Range) SatisfiedBy(v Version) bool {
	if r.Min != nil {
		if !r.Min.Less(v) {
			return false
		}
	}
	if r.Max != nil {
		if !v.Less(*r.Max) {
			return false
		}
	}
	if r.MinEqual != nil {
		if !v.Equals(*r.MinEqual) && !r.MinEqual.Less(v) {
			return false
		}
	}
	if r.MaxEqual != nil {
		if !v.Equals(*r.MaxEqual) && !v.Less(*r.MaxEqual) {
			return false
		}
	}
	return true
}

func (r Range) String() string {
	// Special case for exact equality range
	if r.MinEqual != nil && r.MaxEqual != nil && r.MaxEqual.Equals(*r.MinEqual) {
		return r.MinEqual.String()
	}
	// Special case for tilde and caret ranges
	if r.MinEqual != nil && r.Max != nil {
		if r.Max.Equals(r.MinEqual.IncrementMajor()) {
			return "^" + r.MinEqual.String()
		}
		if r.Max.Equals(r.MinEqual.IncrementMinor()) {
			return "~" + r.MinEqual.String()
		}
	}
	// All other cases
	out := ""
	if r.Min != nil {
		out = ">" + r.Min.String()
	} else if r.MinEqual != nil {
		out = ">=" + r.MinEqual.String()
	}
	if r.Max != nil {
		if out != "" {
			out += " "
		}
		out += "<" + r.Max.String()
	} else if r.MaxEqual != nil {
		if out != "" {
			out += " "
		}
		out += "<=" + r.MaxEqual.String()
	}
	return out
}

func (r Range) Equals(other Range) bool {
	return r.Min.ValueEquals(other.Min) &&
		r.Max.ValueEquals(other.Max) &&
		r.MinEqual.ValueEquals(other.MinEqual) &&
		r.MaxEqual.ValueEquals(other.MaxEqual)
}
