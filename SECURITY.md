# Security policy

## Supported versions

The latest tagged release receives security fixes. Older tags do not.

## Reporting a vulnerability

Report vulnerabilities privately through GitHub's advisory form:

<https://github.com/bmcszk/whenthen/security/advisories/new>

Do not open a public issue for a security report. A good report contains the
version, a reproduction, and the impact you believe it has.

## Disclosure timeline

- **Acknowledgement** within 5 working days of the report.
- **Assessment** of severity and affected versions within 10 working days.
- **Fix and advisory** within 90 days of the report, sooner when the impact is
  severe. You are credited in the advisory unless you ask otherwise.
- **Public disclosure** happens with the advisory, after a fix is released.

If a fix takes longer than 90 days, the advisory is published with mitigation
guidance and the timeline is stated in it.

## Scope

`whenthen` is a static analyzer. It reads Go source and prints diagnostics: it
does not execute analyzed code, makes no network calls, and depends only on
`golang.org/x/tools`.

Realistic vulnerabilities are therefore:

- a crash or panic on valid input, which is a denial of service against your CI,
- the Go plugin entry point loading code it should not,
- supply-chain problems in the published release artifacts.

Release binaries are built in CI and every artifact ships a sigstore
`.sigstore.json` bundle, so a tampered download can be detected.

Disagreements with a lint rule are not vulnerabilities — open a normal issue.
