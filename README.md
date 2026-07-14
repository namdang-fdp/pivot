# Pivot

**Switch projects, keep your flow.**

Pivot is a local project context orchestrator for developers who move among
several projects on one machine. It aims to safely transition the machine from
one complete development context to another while preserving developer data and
leaving resources it does not own untouched.

> **Project status:** Pre-alpha — Slice 0 foundation

## The problem

A development project is more than a command. It can include containers,
native processes, ports, readiness conditions, state, and workspace hooks.
Moving between projects manually makes it easy to leave conflicting resources
running, stop the wrong process, or mistake a started service for a ready one.

Pivot's north-star experience is:

```bash
pivot switch smartpark
```

That command is planned to preflight the target, stop only resources proven to
belong to the active project, start and verify the target, commit truthful state,
and recover safely when a transition fails. It is not implemented in Slice 0.

## Not a Makefile replacement

Make and task runners answer “which commands should run?” Pivot answers “how can
this machine safely enter a complete project context, and how do we prove the
transition succeeded?” Pivot will orchestrate existing tools rather than replace
their build recipes or runtime responsibilities.

## Current CLI

Slice 0 provides build information and help only:

```bash
go run ./cmd/pivot --help
go run ./cmd/pivot version
go run ./cmd/pivot version --json
```

No release artifacts or installation packages are published yet.

## Product pillars

- **Machine awareness:** reconcile recorded intent with what is actually running.
- **Project lifecycle:** provide idempotent, readiness-aware lifecycle operations.
- **Safe context transition:** change projects transactionally and act only with ownership proof.

## Explicit non-goals

Pivot v0.1 is not a runtime installer, process manager, Make replacement,
terminal multiplexer, Kubernetes dashboard, worktree manager, secret manager,
daemon, TUI, or desktop application. It will not allocate ports automatically or
scan repositories with AI.

## Roadmap

Development proceeds in dogfooded vertical slices: repository foundation;
registry, manifest, and diagnostics; single-project lifecycle; ownership and
machine status; transactional switching; resume and workspace hooks; actionable
logs; and release hardening. See [the roadmap](docs/roadmap.md) for boundaries.

## Development

Go 1.25 or later is required.

```bash
go mod download
make fmt
make check
go run ./cmd/pivot version
```

Read [CONTRIBUTING.md](CONTRIBUTING.md) before submitting a change. Product and
architecture details live under [`docs/`](docs/), with the safety rules in
[docs/safety-model.md](docs/safety-model.md).

Security reports must follow [SECURITY.md](SECURITY.md), not a public issue.

## License

Pivot is licensed under the [Apache License 2.0](LICENSE).
