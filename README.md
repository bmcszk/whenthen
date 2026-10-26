# whenthen

`go/analysis` linter: every `func TestX(t *testing.T)` body must mark the when and
then phases, one phase per line, in the order given, when, then.

Two styles count, pick one per test:

```go
// comment style
// given a decoded blob
// when it is decoded
// then the length matches
```

```go
// fluent style, no comments (see the go-fluent-testing skill)
_, when, then := newParts(t)

when.
	createUser()

then.
	noError()
```

Matching is case-insensitive and whole-word. Only phase comments and fluent
receivers count: bare identifiers and string literals never do, so a table field
named `want` or a slice of phase names is ignored.

Alias groups in `Phases` (`whenthen.go`), one word from every group:

| group | optional | aliases |
|-------|----------|---------|
| given | yes | `given`, `arrange`, `setup` |
| when  | no  | `when`, `act` |
| then  | no  | `then`, `assert` |

Messages name the broken rule and carry the `[coding-rules: given/when/then]`
suffix, so a failing gate points at the rule it enforced.

## A marker must own statements

Markers are phases, not headings: each one has to introduce the code of its
phase. A marker glued to the next marker owns nothing and is reported as
`when phase has no statements, a marker must be followed by its code`:

```go
	// when
	// reading constant values to verify enum parity
	// then
	if got := objdat.ThemeCount; got != 17 {
```

That `when` phase holds only a comment, so there is no action to observe. Either
move the marker above real action code, or drop `when` and let the constant
comparison be the `then`.

## Use as a vet tool (no rebuild)

```bash
go install github.com/bmcszk/whenthen/cmd/whenthen@latest
go vet -vettool=$(which whenthen) ./...
```

## Use inside golangci-lint

### Option A: Go plugin (.so, stock binary)

Works with a CGO-enabled golangci-lint and no rebuild, as long as the toolchain
and `golang.org/x/tools` version match the binary's:

```bash
./scripts/build-plugin.sh          # -> ~/.local/lib/golangci/whenthen.so
```

```yaml
version: "2"
linters:
  default: none
  enable: [whenthen]
  settings:
    custom:
      whenthen:
        path: /path/whenthen.so
        description: every Test function needs when and then phases
```

Official release binaries are built with `CGO_ENABLED=0` and fail with
`plugin: not implemented`; use one installed by `go install` (or any
`CGO_ENABLED=1` build) with the same Go version and `x/tools` version as the .so.

### Option B: module plugin (custom binary)

`.custom-gcl.yml`:

```yaml
version: v2.15.0
plugins:
  - module: github.com/bmcszk/whenthen
    import: github.com/bmcszk/whenthen/plugin
    path: /path/to/whenthen   # or drop path and add version: vX.Y.Z for a tagged release
```

```bash
golangci-lint custom          # builds ./custom-gcl
./custom-gcl run ./...
```

`.golangci.yml`:

```yaml
version: "2"
linters:
  default: none
  enable: [whenthen]
  settings:
    custom:
      whenthen:
        type: module
        description: every Test function needs when and then phases
```

The custom binary is pinned to the golangci-lint version in `.custom-gcl.yml`;
rebuild after upgrading.

## Tests

`go test ./...` runs the analyzer against `testdata/src/a`.
