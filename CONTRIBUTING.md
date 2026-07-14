# Contributing to Pivot

Thank you for helping build Pivot. Safety and narrow scope matter as much as
functionality because future commands will control local processes and
containers.

## Requirements and setup

- Go 1.25 or later
- Git
- `golangci-lint` for the full local check (CI also runs it)
- Linux for officially supported v0.1 integration behavior

```bash
git clone <your-fork-url>
cd pivot
go mod download
go test ./...
```

The repository is a modular monolith. Domain concepts belong in `internal/core`,
use-case coordination in `internal/application`, technology-neutral interfaces
in `internal/ports`, infrastructure in `internal/adapters`, and user interfaces
in `internal/presentation`. Dependencies point inward. Read the product,
architecture, safety model, roadmap, and relevant ADRs before coding.

## Development checks

```bash
make fmt
make vet
make test
make test-race
make lint
make build
```

`make check` runs the complete local suite. Keep Go code formatted with `gofmt`.
The lint configuration intentionally favors correctness and lightweight style
checks over noisy complexity metrics.

## Pull requests

Keep each pull request focused on one requested slice or concern. Explain the
user problem, scope, tests, documentation, safety impact, and relevant ADRs.
Include meaningful tests and avoid unrelated refactoring. Conventional Commits
are recommended for readable history, but are not a hard requirement.

Behavior changes require matching documentation. Significant architectural
changes require a new ADR that describes alternatives and consequences. Changes
affecting process termination, container ownership, command execution, state, or
data preservation require explicit positive and negative safety tests.

Do not commit generated binaries, release archives, credentials, local state, or
machine-specific files. A maintainer may request dogfooding evidence for changes
that interact with real runtimes.

By participating, you agree to follow the [Code of Conduct](CODE_OF_CONDUCT.md).
Report exploitable vulnerabilities using [the security policy](SECURITY.md), not
a public issue.
