# whenthen

[![CI](https://github.com/bmcszk/whenthen/actions/workflows/ci.yml/badge.svg)](https://github.com/bmcszk/whenthen/actions/workflows/ci.yml)
[![CodeQL](https://github.com/bmcszk/whenthen/actions/workflows/codeql.yml/badge.svg)](https://github.com/bmcszk/whenthen/actions/workflows/codeql.yml)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/bmcszk/whenthen/badge)](https://securityscorecards.dev/viewer/?uri=github.com/bmcszk/whenthen)
[![Go Reference](https://pkg.go.dev/badge/github.com/bmcszk/whenthen.svg)](https://pkg.go.dev/github.com/bmcszk/whenthen)
[![Go Report Card](https://goreportcard.com/badge/github.com/bmcszk/whenthen)](https://goreportcard.com/report/github.com/bmcszk/whenthen)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A `go/analysis` linter that checks your tests are structured: every
`func TestX(t *testing.T)` has to mark the **when** and **then** phases, one
phase per line, in the order **given → when → then**.

It turns a convention that lives in a document into a gate that fails the build.

## What it reports

| Rule | Message |
|------|---------|
| missing phases | `missing when, then phase(s)` |
| both phases on one line | `when and then on the same line, one phase per line` |
| wrong order | `then before when, want given, when, then` |
| mandatory marker with no code | `when phase has no statements, a marker must be followed by its code` |
| optional marker with no code | `given marker has no statements, given is optional: delete the marker or move the setup under it` |
| subtest breaks a rule | `TestX: subtest at line 12: missing then phase(s)` |

Every message ends with ` [coding-rules: given/when/then]` and starts with the
test name, so a red build tells you which test and which rule:

```
internal/assets/cel_test.go:18:1: TestDecodeCEL: when and then on the same line, one phase per line [coding-rules: given/when/then] (whenthen)
```

## Install

### As a vet tool (simplest)

```bash
go install github.com/bmcszk/whenthen/cmd/whenthen@latest
go vet -vettool=$(which whenthen) ./...
```

Exit code is non-zero when a test breaks a rule, so it drops straight into any
CI step or `justfile` recipe.

### Inside golangci-lint

Three integration paths exist; pick the one that matches how you install
golangci-lint.

**1. Go plugin (`.so`)** — stock golangci-lint, no rebuild:

```bash
git clone https://github.com/bmcszk/whenthen && cd whenthen
./scripts/build-plugin.sh      # -> ~/.local/lib/golangci/whenthen.so
```

```yaml
version: "2"
linters:
  default: none
  enable: [whenthen]
  settings:
    custom:
      whenthen:
        path: /home/you/.local/lib/golangci/whenthen.so
        description: tests must mark given, when and then phases
```

The `.so` only loads into a CGO-enabled golangci-lint built with the same Go
version and the same `golang.org/x/tools` version. Official release binaries are
built with `CGO_ENABLED=0` and answer `plugin: not implemented`; use a binary
from `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`.

**2. Module plugin (custom binary)** — for teams pinned to a released version:

```yaml
# .custom-gcl.yml
version: v2.15.0
plugins:
  - module: github.com/bmcszk/whenthen
    import: github.com/bmcszk/whenthen/plugin
    version: v0.1.0
```

```bash
golangci-lint custom   # builds ./custom-gcl
./custom-gcl run ./...
```

Enable it in `.golangci.yml` with `settings.custom.whenthen.type: module`.

**3. `go vet -vettool`** — no golangci-lint at all, see above.

## What counts as a phase

Two styles are accepted. Use one per test.

**Comment markers** — a comment line naming the phase, above the code it
governs:

```go
func TestDecodeTRN(t *testing.T) {
	// given three TRN blobs
	blobs := loadBlobs(t)

	// when each blob is decoded
	frames := decodeAll(t, blobs)

	// then the frame counts and data lengths match
	requireFrameCounts(t, frames, want)
}
```

**Fluent chains** — the phase is the receiver of the chain, so no comments at
all:

```go
func TestCreateUser(t *testing.T) {
	_, when, then := newParts(t)

	when.
		createUser("bender")

	then.
		noError()
}
```

Matching is case-insensitive and whole-word. Only phase comments and fluent
chain receivers count — bare identifiers and string literals are ignored, so a
table field named `want` or a slice of phase names never triggers a false
positive.

**A marker opens the line.** `// when the blob is decoded` is a marker;
`// if the blob is empty then skip` is prose that happens to contain a phase
word, so it never counts. That keeps ordinary English out of the gate.

### Subtests

Markers belong to the function body they sit in, so a `t.Run` closure is judged
on its own statements instead of the test body's:

```go
func TestDecode(t *testing.T) {
	for _, name := range names() {
		t.Run(name, func(t *testing.T) {
			// given a raw name
			raw := name

			// when it is decorated
			out := decorate(raw)

			// then it stays loud
			if out == "" {
				t.Fatalf("%s stayed quiet", name)
			}
		})
	}
}
```

That test passes without markers of its own: the phases are owned by the
closure, and a test that only fans out to subtests leaves the phases to them. A
closure *without* markers is a step of its test and is left alone. A subtest that
breaks a rule is reported on its own closing brace:

```
decode_test.go:12:2: TestDecode: subtest at line 14: missing then phase(s) [coding-rules: given/when/then]
```

### Aliases

One word from each group is required:

| Phase | Optional | Aliases |
|-------|----------|---------|
| given | yes | `given`, `arrange`, `setup` |
| when | no | `when`, `act` |
| then | no | `then`, `assert` |

`given` is optional because setup is not always worth a line — a test can start
straight at `when`. `when` and `then` are mandatory: a test without an action or
without an assertion is not a test.

### A marker must own statements

A marker names the code right below it, so a marker with no statement under it is
decoration. The message tells you which fix applies:

* `when phase has no statements, a marker must be followed by its code` — the
  phase is mandatory, so write the action or assertion it promised.
* `given marker has no statements, given is optional: delete the marker or move
  the setup under it` — `given` is optional, so the usual fix is to drop the
  marker instead of inventing setup code for it.

That keeps agents from "fixing" a dangling `given` by shuffling the test around
when deleting one line is the whole repair.

## Why `given` is optional but nothing else is

A test without `when` only exercises a fixture. A test without `then` asserts
nothing. Both are already broken; the linter just says so before review does.

## FAQ

**Does it check that my assertions are meaningful?** No. It checks structure, not
substance — that the phases exist, in order, on their own lines, each owning
code.

**Why not accept `want` as a `then` alias?** `want` is the Go convention for the
expected value (`// got, want`), not a phase marker. Accepting it would mark
table-driven tests as structured when nothing is structured.

**What about `t.Run` subtests?** Only `func TestX(t *testing.T)` top-level
functions are checked. The body can contain any structure you like, including
subtests — a dispatcher that only calls `t.Run` still needs its own markers.

**Can I disable one rule?** No. Rules are not individually toggleable: the point
is the whole shape. Use `//nolint:whenthen` with a reason if you must, and expect
review to ask why.

## Development

Architecture, local gates, testing strategy and the release process live in
[CONTRIBUTING.md](CONTRIBUTING.md). Security reports: [SECURITY.md](SECURITY.md).

## License

[MIT](LICENSE) © 2026 bmcszk
