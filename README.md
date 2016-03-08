# semv

semv is a semantic versioning (semver) library for go

## Ethos

This library is designed to be friendly to common use-cases of semantic versioning, and follows [the semver v2.0.0 spec] closely, whilst being permissive to additional real-world usages of semver-like versions, for example partial versions like `1` or `5.3`. Whilst not valid semver versions, these are commonly seen in the wild, and I think a good library should allow us to work with these kinds of versions as well.

If you really care that only exact semver v2.0.0 versions are used, you can use `.ParseExact(Semver2_0_0)` which will error if the string does not follow the spec.

[the semver v2.0.0 spec]: http://semver.org/spec/v2.0.0.html

### Parsing

Parsing using `Parse()` is permissive by default. All of the following will parse into a Version instance:

- `"1"` parses as `{Major: 1, Minor: 0, Patch: 0, Pre: "", Meta: ""}`
- `"1.2"` parses as `{Major: 1, Minor: 2, Patch: 0, Pre: "", Meta: ""}`
- `"1.2.3"` parses as `{Major: 1, Minor: 2, Patch: 3, Pre: "", Meta: ""}`
- `"1-beta"` parses as `{Major: 1, Minor: 0, Patch: 0, Pre: "beta", Meta: ""}`
- `"1.2+abc"` parses as `{Major: 1, Minor: 2, Patch: 0, Pre: "", Meta: "abc"}`

### String()

Simply calling `.String()` on a version created using one of the `New(Version|MMP)` funcs will print the full version string, ommitting the optional prerelease and/or metadata sections depending on if they contain any data.

If a version is created by parsing a string, its original format is recorded with the version. In this case, calling `.String()` will print the version it its original format. E.g.:

- `Parse("1").String() == "1"`
- `Parse("1.0").String() == "1.0"`
- `Parse("1.0.0").String() == "1.0.0"`
- `Parse("1.0.0-beta").String() == "1.0.0-beta"`
- `Parse("1.0.0-beta+abc").String() == "1.0.0-beta+abc"`
- `Parse("1.0.0+abc").String() == "1.0.0+abc"`

This feature is useful when dealing with partial versions, as used by many popular projects, e.g. [Go], [NPM], and others. You can always use `.Format()` to print an exact format.

[Go]: https://golang.org
[NPM]: https://www.npmjs.com

### Format()

If you want to print your version in a specific format, you can use `.Format()` with a format string, e.g.:

- `Parse("1.2.3-beta.1+abc-def.2").Format("M") == "1"`
- `Parse("1.2.3-beta.1+abc-def.2").Format("M.m") == "1.2"`
- `Parse("1.2.3-beta.1+abc-def.2").Format("M.m.p") == "1.2.3"`
- `Parse("1.2.3-beta.1+abc-def.2").Format("M.m.p-?") == "1.2.3-beta.1"`
- `Parse("1.2.3-beta.1+abc-def.2").Format("M.m.p+?") == "1.2.3+abc-def.2"`
- `Parse("1.2.3-beta.1+abc-def.2").Format("M.m.p-?+?") == "1.2.3-beta.1+abc-def.2"`


