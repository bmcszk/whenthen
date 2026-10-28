# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `examples/golangci-full.yml`: the complete, working config from a production
  repository, copied verbatim, with whenthen enabled next to thirty-two other
  linters. The minimal examples show how to wire one linter; this one shows what
  a real config looks like, including per-linter settings and the exclusion that
  keeps `errcheck` quiet in tests.

### Changed

- The README now presents the integration paths in the order upstream recommends
  them (module plugin first, Go plugin second), notes that
  `golangci/golangci-lint-action` v9 builds `.custom-gcl.yml` on its own, and
  documents the Go 1.24 `go get -tool` route that needs no config at all.

## [0.1.2] - 2026-10-27

### Added

- Subtests are scopes of their own: a `t.Run` closure is judged on its own
  statements, so markers inside a closure no longer claim the test body's phase
  lines. A test that only fans out to subtests leaves the phases to them, and a
  broken subtest is reported as `TestX: subtest at line N: <rule>`.
- A fuzz corpus entry for the invalid-UTF-8 comment case, so the panic below
  cannot come back unnoticed.

### Changed

- The empty-marker message now depends on the phase. A mandatory phase still says
  `when phase has no statements, a marker must be followed by its code`, while an
  optional one says `given marker has no statements, given is optional: delete the
  marker or move the setup under it`, so the reported fix is the one that applies.

### Fixed

- A marker now has to open the comment line. Prose that merely contains a phase
  word, such as `// if the blob is empty then skip`, is no longer read as a `then`
  marker.
- `hasWord` no longer panics on comment text with invalid UTF-8. It used to feed
  the bytes straight into `regexp.Compile`; the nightly fuzz job found it.

## [0.1.1] - 2026-10-27

Documentation and packaging only: README rewritten for end users, consumer
integration examples, `.gitattributes` so fixtures diff as Go.

## [0.1.0] - 2026-10-26

### Added

- `whenthen` analyzer: every `func TestX(t *testing.T)` must mark the given, when
  and then phases, one phase per line, in that order.
- Phase alias groups: `given`/`arrange`/`setup`, `when`/`act`, `then`/`assert`.
  `given` is optional; `when` and `then` are mandatory.
- Rules: missing phase, two phases on one line, out-of-order phases, and a marker
  that owns no statements.
- Comment markers (`// when`) and fluent chain receivers (`when.act()`) are both
  accepted; bare identifiers and string literals are ignored.
- Diagnostics carry the test name, the broken rule and the
  `[coding-rules: given/when/then]` hint.
- Integrations: `go vet -vettool`, stock golangci-lint via Go plugin (`.so`), and
  golangci-lint module plugin (`golangci-lint custom`).
- `analysistest` fixture suite plus white-box table tests, 93% coverage.
- Self-check: the repository lints its own tests with the analyzer.

[Unreleased]: https://github.com/bmcszk/whenthen/compare/v0.1.2...HEAD
[0.1.2]: https://github.com/bmcszk/whenthen/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/bmcszk/whenthen/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/bmcszk/whenthen/releases/tag/v0.1.0
