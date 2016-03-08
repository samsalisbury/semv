package semv

type (
	// Range is a semver range.
	Range struct {
		GreaterThan, LessThan *Version
	}
)
