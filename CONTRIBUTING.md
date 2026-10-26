# Contributing

## Repo layout

```
whenthen.go              the analyzer: phase table, marker detection, rules
whenthen_test.go         white-box tests (tables, 90%+ coverage)
testdata/src/a/a.go      analysistest fixtures, one want comment per violation
cmd/whenthen/main.go     singlechecker binary, used by go vet -vettool
plugin/plugin.go         golangci-lint module-plugin wrapper
pluginmain/plugin.go     golangci-lint Go plugin (buildmode=plugin) entry point
scripts/build-plugin.sh  builds the .so into $WHENTHEN_PLUGIN_DIR
justfile                 local gates: build, test, cover, lint, selfcheck, verify
```

## How the analyzer works

1. `isTest` accepts a top-level `func TestX(t *testing.T)` with a body. Methods,
   helpers, other parameter types and `testing.B` are skipped.
2. `bodyMarkers` collects markers from two sources: comment lines inside the body
   that name a phase (whole-word, case-insensitive), and `ast.SelectorExpr`
   receivers named after a phase (`when.act()`). Bare identifiers and literals are
   deliberately ignored — that is what keeps `want`, variable names and slices of
   phase names out of the result.
3. `locatePhases` keeps the earliest line per phase.
4. `violation` applies the rules in order and returns the first failure:
   missing phases → same line → wrong order → marker with no statements.
   `stmtSpans` + `emptyPhase` implement the last one: a marker owns the
   top-level statements between itself and the next marker.

Adding a rule means adding a function with a table test and a fixture, in that
order of preference. Add the fixture to `testdata/src/a/a.go` with a `// want`
comment on the closing brace line — the comment must sit outside the body,
otherwise it is parsed as a phase marker.

## Local gates

```bash
just verify     # fmt check + lint + selfcheck + test
just test       # go test ./...
just cover      # coverage profile
just selfcheck  # dogfood: build the analyzer and vet this repo with it
just plugin     # build ~/.local/lib/golangci/whenthen.so
```

`just verify` is what CI runs. It must be green before you push; a red gate is a
bug in the change, not in the gate.

The self-check (`just selfcheck`) runs the analyzer over this repository's own
tests. If you add a test, it needs phase markers too — the linter is not allowed
to be exempt from itself.

## Testing strategy

- White-box tables for pure functions (`phaseLine`, `hasStatement`, `isTest`,
  `hasWord`, `emptyPhase`), with the `given`/`when`/`then` markers written as real
  phase comments in the tests themselves.
- `analysistest` fixtures for end-to-end behaviour: every rule has a fixture that
  must fire and at least one that must stay silent (aliases, fluent chains,
  identifier-only lookalikes).
- Coverage is kept above 90%; the uncovered remainder is error paths in the
  helper parser.

## golangci-lint integration

Two adapters exist and both must compile after an API change:

- `pluginmain/plugin.go` — `package main` with `func New(any) ([]*analysis.Analyzer, error)`
  and an empty `func main()`, built with `-buildmode=plugin`. Needs CGO and a
  golangci-lint built by the same toolchain.
- `plugin/plugin.go` — module plugin via `github.com/golangci/plugin-module-register`.

After touching the analyzer, rebuild the plugin so local projects pick it up:

```bash
just plugin
```

## Releasing

1. Update `CHANGELOG.md`.
2. Tag: `git tag vX.Y.Z && git push origin vX.Y.Z`.
3. The release workflow builds linux/darwin/windows binaries for amd64 and arm64,
   attaches them to the release and generates notes.

Versioning is semantic: new rule or new alias group = minor, message wording or
internal refactor = patch, behaviour removed = major.

## Commits and PRs

- One concern per commit; the message says what changed in the gate, not what
  files were touched.
- Every PR must keep `just verify` green and add a fixture for new behaviour.
- Tests that assert structure go with the code; do not land the rule without the
  fixture that proves it fires.
