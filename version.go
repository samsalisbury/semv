package semv

import (
	"fmt"
	"strconv"
	"strings"
)

type (
	// Version is a semver version
	Version struct {
		Major, Minor, Patch      int
		Pre, Meta, DefaultFormat string
	}
	// VersionIncomplete is an error returned by ParseExactSemver2_0_0
	// when a version is missing either minor or patch parts.
	VersionIncomplete struct {
		MissingPart string
	}
	// UnexpectedCharacter is an error returned by Parse and ParseExactSemver2_0_0
	// when they contain unexpected characters at a particular location.
	UnexpectedCharacter struct {
		Char rune
		Pos  int
	}
	// ZeroLengthNumeric is an error returned when either major, minor, or
	// patch is zero length. That is, when parsing a string containing two
	// consecutive dots. E.g. "1..3" or "..1"
	ZeroLengthNumeric struct {
		ZeroLengthPart string
	}
	// PrecedingZero is an error returned when one of the major, minor, or
	// patch parts contains a preceding zero. This error is only returned
	// when using ParseExactSemver2_0_0, and this validation is ignored
	// otherwise.
	PrecedingZero struct {
		PrecedingZeroPart, InputString string
	}
	mode uint
)

func NewVersion(major, minor, patch int, pre, meta string) Version {
	return Version{major, minor, patch, pre, meta, ""}
}

func NewMajorMinorPatch(major, minor, patch int) Version {
	return Version{major, minor, patch, "", "", ""}
}

func (err VersionIncomplete) Error() string {
	return fmt.Sprintf("version incomplete: missing %s component", err.MissingPart)
}

func (err UnexpectedCharacter) Error() string {
	return fmt.Sprintf("unexpected character '%s' at position %d", string(err.Char), err.Pos)
}

func (err ZeroLengthNumeric) Error() string {
	return fmt.Sprintf("unexpected zero-length %s component", err.ZeroLengthPart)
}

func (err PrecedingZero) Error() string {
	return fmt.Sprintf("unexpected preceding zero in %s component: %q",
		err.PrecedingZeroPart, err.InputString)
}

const (
	modeMajor            mode = iota
	modeMinor                 = iota
	modePatch                 = iota
	modePre                   = iota
	modeMeta                  = iota
	digits                    = "01234567890"
	validPreAndMetaChars      = digits + ".-abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Major                     = "M"
	Minor                     = "m"
	Patch                     = "p"
	PreDelim                  = "-"
	Pre                       = PreDelim + "?"
	PreRaw                    = PreDelim + "!"
	MetaDelim                 = "+"
	Meta                      = MetaDelim + "?"
	MetaRaw                   = MetaDelim + "!"
	MajorMinor                = Major + "." + Minor
	MajorMinorPatch           = MajorMinor + "." + Patch
	MMPPre                    = MajorMinorPatch + Pre
	Complete                  = MMPPre + Meta
	Semver_2_0_0              = Complete
)

func (v Version) Validate() error {
	if v.Major < 0 || v.Minor < 0 || v.Patch < 0 {
		return fmt.Errorf("major, minor, patch must all be positive")
	}
	return nil
}

// String prints the string representation of this version.
// If the version was originally parsed, then String will attempt
// to re-print it at the same level of detail as was originally parsed in.
//
// E.g. Parse("1").String() == "1"
//      Parse("1.2").String() == "1.2"
//      Parse("1.2.3").String() == "1.2.3"
//      Parse("1.2.3-beta").String() == "1.2.3-beta"
func (v Version) String() string {
	return v.Format(v.DefaultFormat)
}

// Format takes a format string and outputs the version accordingly.
//
// You can use the following format strings (which are available as constants):
//
//     Major = "M", minor = "m", Patch = "p", Pre = "-?", Meta = "+?",
//     PreRaw = "-!", MetaRaw = "+!"
//
// Pre and Meta are replaced with the empty string when Pre or Meta are empty,
// respectively, or, with the prerelease version prefixed by '-' or the metadata
// prefixed with '+', if either are not empty.
//
// See other constants in this library for more. The empty string is treated
// equivalently to the format string "M.m.p-?+?".
func (v Version) Format(format string) string {
	if format == "" {
		format = Complete
	}
	replacements := map[string]interface{}{
		Major: v.Major,
		Minor: v.Minor,
		Patch: v.Patch,
	}
	formatted := replaceAll(format, replacements)
	if v.Pre != "" {
		formatted = strings.Replace(formatted, Pre, PreDelim+v.Pre, -1)
	} else {
		formatted = strings.Replace(formatted, Pre, "", -1)
	}
	if v.Meta != "" {
		formatted = strings.Replace(formatted, Meta, MetaDelim+v.Meta, -1)
	} else {
		formatted = strings.Replace(formatted, Meta, "", -1)
	}
	formatted = strings.Replace(formatted, PreRaw, v.Pre, -1)
	formatted = strings.Replace(formatted, MetaRaw, v.Meta, -1)
	return formatted
}

func (v Version) Less(than Version) bool {
	if v.Major < than.Major {
		return true
	}
	if v.Major > than.Major {
		return false
	}
	if v.Minor < than.Minor {
		return true
	}
	if v.Minor > than.Minor {
		return false
	}
	if v.Patch < than.Patch {
		return true
	}
	if v.Patch > than.Patch {
		return false
	}
	if v.Pre != "" && than.Pre == "" {
		return true
	}
	if v.Pre == "" && than.Pre != "" {
		return false
	}
	vPreParts := v.PreComponents()
	thanPreParts := than.PreComponents()
	for i := range thanPreParts {
		if vInt, err := strconv.Atoi(vPreParts[i]); err == nil {
			if tInt, err := strconv.Atoi(thanPreParts[i]); err == nil {
				if vInt < tInt {
					return true
				}
				if vInt > tInt {
					return false
				}
			} else {
				return true
			}
		}
		if vPreParts[i] < thanPreParts[i] {
			return true
		}
		if vPreParts[i] > thanPreParts[i] {
			return false
		}
		if len(vPreParts) == i+1 && len(thanPreParts) > len(vPreParts) {
			return true
		}
	}
	return false
}

func (v Version) PreComponents() []string {
	return strings.Split(v.Pre, ".")
}

func (v Version) Equals(other Version) bool {
	return !v.Less(other) && !other.Less(v)
}

// ValueEquals works on Version pointers, and checks that their values
// are equal. Also returns true if both are nil, and obviously false if
// one is nil and ther other one isn't.
func (v *Version) ValueEquals(other *Version) bool {
	if v == nil && other == nil {
		return true
	}
	if v == nil || other == nil {
		return false
	}
	return (*v).Equals(*other)
}

func (v Version) Satisfies(r Range) bool {
	return r.SatisfiedBy(v)
}

func (v Version) IncrementMajor() Version {
	v.Major++
	v.Minor = 0
	v.Patch = 0
	return v
}

func (v Version) IncrementMinor() Version {
	v.Minor++
	v.Patch = 0
	return v
}

func (v Version) IncrementPatch() Version {
	v.Patch++
	return v
}
