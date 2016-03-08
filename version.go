package semv

import (
	"bytes"
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
	// Range is a semver range
	Range struct {
		GreaterThan, LessThan Version
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
	return fmt.Sprintf("version incomplete: missing %s", err.MissingPart)
}

func (err UnexpectedCharacter) Error() string {
	return fmt.Sprintf("unexpected character '%s' at position %d", string(err.Char), err.Pos)
}

func (err ZeroLengthNumeric) Error() string {
	return fmt.Sprintf("unexpected zero-length %s", err.ZeroLengthPart)
}

func (err PrecedingZero) Error() string {
	return fmt.Sprintf("unexpected preceding zero on %s: %q",
		err.PrecedingZeroPart, err.InputString)
}

// Parse permissively parses the string as a semver value. The minimal string
// which will not error is a single digit, which will be interpreted as a major
// version, e.g. Parse("1").Format("M.m.p") == "1.0.0".
func Parse(s string) (Version, error) {
	v, err := parse(s)
	if err == nil {
		return v, nil
	}
	if _, ok := err.(VersionIncomplete); ok {
		return v, nil
	}
	return v, err
}

// ParseExactSemver2_0_0 returns an error, and an incomplete Version if the
// string passed in does not conform exactly to semver 2.0.0
func ParseExactSemver2_0_0(s string) (Version, error) {
	return parse(s)
}

// ParseAny tries to parse any version found in a string. It starts
// parsing at the first decimal digit [0-9], and stops when it finds
// an invalid character. It returns an error only if there are no
// digits found in the string.
func ParseAny(s string) (Version, error) {
	startIndex := strings.IndexAny(s, digits)
	if startIndex == -1 {
		return Version{}, fmt.Errorf("no version found in %q", s)
	}
	v, _ := Parse(s[startIndex:])
	return v, nil
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

func parse(s string) (Version, error) {
	var parsedMinor, parsedPatch, parsedPre, parsedMeta bool
	var (
		major = &bytes.Buffer{}
		minor = &bytes.Buffer{}
		patch = &bytes.Buffer{}
		pre   = &bytes.Buffer{}
		meta  = &bytes.Buffer{}
	)
	targets := map[mode]*bytes.Buffer{
		modeMajor: major,
		modeMinor: minor,
		modePatch: patch,
		modePre:   pre,
		modeMeta:  meta,
	}
	m := modeMajor
	var i int
	var c rune
	// finalise takes the current buffers and tries to return a partial version
	finalise := func(knownErrors ...error) (Version, error) {
		var err error
		v := Version{}
		v.DefaultFormat = Major
		majorString := major.String()
		if v.Major, err = strconv.Atoi(majorString); err != nil {
			return v, firstErr(append(knownErrors, err)...)
		}
		if err := validateMMPFormat(majorString, "major"); err != nil {
			knownErrors = append(knownErrors, err)
		}
		if parsedMinor {
			v.DefaultFormat = MajorMinor
			minorString := minor.String()
			if v.Minor, err = strconv.Atoi(minorString); err != nil {
				return v, firstErr(append(knownErrors, err)...)
			}
			if err := validateMMPFormat(minorString, "minor"); err != nil {
				knownErrors = append(knownErrors, err)
			}
		}
		if parsedPatch {
			v.DefaultFormat = MajorMinorPatch
			patchString := patch.String()
			if v.Patch, err = strconv.Atoi(patchString); err != nil {
				return v, firstErr(append(knownErrors, err)...)
			}
			if err := validateMMPFormat(patchString, "patch"); err != nil {
				knownErrors = append(knownErrors, err)
			}
		}
		if parsedPre {
			v.DefaultFormat = v.DefaultFormat + "-?"
		}
		if parsedMeta {
			v.DefaultFormat = v.DefaultFormat + "+?"
		}
		v.Pre = pre.String()
		v.Meta = meta.String()
		return v, firstErr(append([]error{v.Validate()}, knownErrors...)...)
	}
	changeMode := func() (bool, error) {
		if (m == modePre || m == modeMeta) && c == '-' {
			return false, nil
		}
		if m == modeMeta && c == '+' {
			return false, UnexpectedCharacter{c, i}
		}
		if m == modePatch && c == '.' {
			return false, UnexpectedCharacter{c, i}
		}
		if (m == modeMajor || m == modeMinor) && c == '.' {
			m++
			return true, nil
		}
		switch c {
		default:
			return false, nil
		case '-':
			m = modePre
		case '+':
			m = modeMeta
		}
		return true, nil
	}
	for i, c = range s {
		if m == modeMinor {
			parsedMinor = true
		}
		if m == modePatch {
			parsedPatch = true
		}
		if m == modePre {
			parsedPre = true
		}
		if m == modeMeta {
			parsedMeta = true
		}
		switch c {
		case '.', '-', '+':
			changed, err := changeMode()
			if err != nil {
				return finalise(err)
			}
			if changed {
				continue
			}
		}
		switch m {
		case modeMajor, modeMinor, modePatch:
			if strings.ContainsRune(digits, c) {
				targets[m].WriteRune(c)
			} else {
				return finalise(UnexpectedCharacter{c, i})
			}
		case modePre, modeMeta:
			if strings.ContainsRune(validPreAndMetaChars, c) {
				targets[m].WriteRune(c)
			} else {
				return finalise(UnexpectedCharacter{c, i})
			}
		}
	}
	if !parsedMinor {
		return finalise(VersionIncomplete{"minor"})
	}
	if !parsedPatch {
		return finalise(VersionIncomplete{"patch"})
	}
	return finalise(nil)
}

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

func replaceAll(s string, replacements map[string]interface{}) string {
	for what, replacement := range replacements {
		s = strings.Replace(s, what, fmt.Sprint(replacement), -1)
	}
	return s
}

func firstErr(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func validateMMPFormat(s, name string) error {
	if len(s) == 0 {
		return ZeroLengthNumeric{name}
	}
	if len(s) > 1 && s[0] == '0' {
		return PrecedingZero{name, s}
	}
	return nil
}
