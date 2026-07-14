# Security policy

## Supported versions

Pivot is pre-alpha and has no stable supported release.

| Version | Supported |
| --- | --- |
| Stable releases | None yet |
| `main` | Best-effort development fixes |

## Reporting a vulnerability

Do not open a public issue for an exploitable vulnerability. Use GitHub private
vulnerability reporting for this repository when it is available. Otherwise,
contact the maintainers privately through the repository owner's published
GitHub channels. Do not include secrets or sensitive machine data in an initial
message.

Please describe the affected revision, impact, reproduction conditions, and any
known mitigations. Redact credentials, tokens, personal paths, and unrelated
process or container data. Maintainers will acknowledge the report, investigate,
and coordinate disclosure based on severity and available project capacity.

## Security-sensitive areas

Extra scrutiny is required for:

- process identity, signaling, and termination, including PID reuse;
- container and project-namespace ownership;
- path traversal, symlinks, and filesystem permissions;
- external command construction and argument boundaries;
- environment inheritance, expansion, logging, and secret redaction;
- state-file locking, integrity, permissions, and atomic replacement.

Pivot's governing rule is: no ownership proof, no destructive action. Reports
that show a way to cross project ownership boundaries or lose developer data are
especially important.
