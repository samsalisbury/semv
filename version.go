package semv

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

type Version struct {
	Major, Minor, Patch      int
	Pre, Meta, DefaultFormat string
}

type Range struct {
	GreaterThan, LessThan Version
}

func NewVersion(major, minor, patch int, pre, meta string) Version {
	return Version{major, minor, patch, pre, meta, ""}
}

func NewMajorMinorPatch(major, minor, patch int) Version {
	return Version{major, minor, patch, "", "", ""}
}

type mode uint

const (
	modeMajor       mode = iota
	modeMinor            = iota
	modePatch            = iota
	modePre              = iota
	modeMeta             = iota
	digits               = "01234567890"
	Major                = "M"
	Minor                = "m"
	Patch                = "p"
	PreDelim             = "-"
	Pre                  = PreDelim + "?"
	PreRaw               = PreDelim + "!"
	MetaDelim            = "+"
	Meta                 = MetaDelim + "?"
	MetaRaw              = MetaDelim + "!"
	MajorMinor           = Major + "." + Minor
	MajorMinorPatch      = MajorMinor + "." + Patch
	MMPPre               = MajorMinorPatch + Pre
	Complete             = MMPPre + Meta
)

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

func Parse(s string) (Version, error) {
	var parsedMinor, parsedPatch bool
	var (
		v     Version
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
	unexpectedCharacter := func() error {
		return fmt.Errorf("unexpected character '%s' at position %d", string(c), i)
	}
	changeMode := func() (bool, error) {
		if (m == modePre || m == modeMeta) && c == '-' {
			return false, nil
		}
		if m == modeMeta && c == '+' {
			return false, unexpectedCharacter()
		}
		if m == modePatch && c == '.' {
			return false, unexpectedCharacter()
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
		switch c {
		case '.', '-', '+':
			changed, err := changeMode()
			if err != nil {
				return v, err
			}
			if changed {
				continue
			}
		}
		switch m {
		case modeMajor, modeMinor, modePatch:
			switch c {
			default:
				return v, unexpectedCharacter()
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				targets[m].WriteRune(c)
			}
		case modePre, modeMeta:
			targets[m].WriteRune(c)
		}
	}
	var err error
	v.DefaultFormat = Major
	if v.Major, err = strconv.Atoi(major.String()); err != nil {
		return v, err
	}
	if parsedMinor {
		v.DefaultFormat = MajorMinor
		if v.Minor, err = strconv.Atoi(minor.String()); err != nil {
			return v, err
		}
	}
	if parsedPatch {
		v.DefaultFormat = ""
		if v.Patch, err = strconv.Atoi(patch.String()); err != nil {
			return v, err
		}
	}
	v.Pre = pre.String()
	v.Meta = meta.String()
	return v, v.Validate()
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
