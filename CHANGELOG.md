# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/bmcszk/whenthen/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/bmcszk/whenthen/releases/tag/v0.1.0
