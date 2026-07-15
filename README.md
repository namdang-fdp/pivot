# Pivot

**Switch projects, keep your flow.**

Pivot is a local project context orchestrator for developers who move among
several projects on one machine. It aims to safely transition the machine from
one complete development context to another while preserving developer data and
leaving resources it does not own untouched.

> **Project status:** Pre-alpha — Slice 1: registry, manifest, and doctor

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
and recover safely when a transition fails. Transactional switching remains
planned and is not implemented in Slice 1.

## Not a Makefile replacement

Make and task runners answer “which commands should run?” Pivot answers “how can
this machine safely enter a complete project context, and how do we prove the
transition succeeded?” Pivot will orchestrate existing tools rather than replace
their build recipes or runtime responsibilities.

## Current CLI

Slice 1 implements explicit project manifest creation, registration, listing,
and read-only machine diagnostics:

```bash
go run ./cmd/pivot --help
go run ./cmd/pivot version
go run ./cmd/pivot version --json
go run ./cmd/pivot init [path]
go run ./cmd/pivot add [path]
go run ./cmd/pivot list
go run ./cmd/pivot list --json
go run ./cmd/pivot doctor [project]
go run ./cmd/pivot doctor [project] --json
```

`init`, `add`, `list`, and `doctor` do not start, stop, or modify development
resources. Doctor is inspection-only. See the [command reference](docs/commands.md)
and [manifest reference](docs/manifest-reference.md).

The project registry is stored at `$XDG_CONFIG_HOME/pivot/projects.yaml`, with a
fallback of `~/.config/pivot/projects.yaml`. Isolate experiments and tests with:

```bash
XDG_CONFIG_HOME=<temp-directory> go run ./cmd/pivot list
```

No release artifacts or installation packages are published yet.

## Planned commands

The lifecycle commands `up`, `down`, `status`, `switch`, `resume`, and `shutdown`
are planned for later slices and are not currently usable.

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

Development proceeds in dogfooded vertical slices. The repository foundation and
registry/manifest/doctor implementation precede single-project lifecycle,
ownership and machine status, transactional switching, resume and workspace
hooks, actionable logs, and release hardening. See
[the roadmap](docs/roadmap.md) for boundaries.

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
