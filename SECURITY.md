# Security policy

## Supported versions

The latest tagged release receives fixes. Older tags do not.

## Reporting a vulnerability

Report privately via GitHub's advisory form:
<https://github.com/bmcszk/whenthen/security/advisories/new>

Do not open a public issue for a security report. Include the version, a
reproduction and the impact you believe it has. Expect an initial reply within a
week.

## Scope

`whenthen` is a static analyzer. It reads Go source and prints diagnostics; it
does not execute analyzed code, does not make network calls and has no runtime
dependencies beyond `golang.org/x/tools`.

Realistic issues are therefore limited to:

- analyzer crashes or panics on valid input (a denial of service against your CI),
- the Go plugin entry point loading code it should not,
- supply-chain problems in the release artifacts.

Anything else — including disagreements with a lint rule — belongs in a normal
issue.
