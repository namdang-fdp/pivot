# Instructions for coding agents

Pivot coordinates destructive-capable operations on a developer's machine. Scope
and safety rules are mandatory. Never trade them for speed or convenience.

## Read before coding

Before changing code, read `docs/product.md`, `docs/philosophy.md`,
`docs/specification-v0.1.md`, `docs/architecture.md`, `docs/safety-model.md`, and
the ADRs relevant to the requested work. Read the roadmap entry for the active
slice. If requirements conflict, stop and report the conflict rather than
silently choosing a broader interpretation.

## Scope discipline

- Work only within the explicitly requested slice. Never silently expand scope.
- Do not implement a daemon, Kubernetes support, or login startup in v0.1.
- Do not add TUI or desktop dependencies without a dedicated milestone.
- Do not build a custom process scheduler; Process Compose is the intended v0.1
  native process backend.
- Do not create public packages or a `pkg/` tree without an accepted ADR and a
  deliberate versioned SDK contract.
- Prefer explicit configuration and enrollment over repository auto-detection.
- Avoid speculative interfaces and abstractions without a current use case.

Apply this permanent product check to every change:

> Does this change improve machine awareness, project lifecycle, or safe context transition?

If the answer is no, it probably does not belong in Pivot's current scope.

## Safety invariants

- Never bypass ownership verification. No ownership proof means no destructive
  action.
- Never terminate processes by generic name. Do not use `pkill`, `killall`, or a
  stored PID without reconciled identity evidence.
- Never stop unmanaged containers or processes.
- Never delete Docker volumes by default.
- Never introduce destructive cleanup commands in v0.1.
- Preserve truthful observed state after failures; do not claim rollback or
  readiness that was not proven.
- Add explicit tests for lifecycle, ownership, failure, interruption, rollback,
  and data-preservation behavior whenever those areas are changed.

## Architecture boundaries

- Keep business logic out of Cobra handlers and `main.go`.
- Core packages must not depend on presentation, infrastructure, Cobra, Wails,
  YAML, Docker, Process Compose, shell execution, `$HOME`, or terminal output.
- Application services coordinate use cases through core concepts and ports; they
  do not build Docker-specific commands.
- External processes and filesystems belong in adapters.
- CLI, future TUI, and future desktop are peer presentation layers calling the
  same application services. Never make the desktop execute the CLI binary.
- Use standard `log/slog` for future diagnostic logging; keep user-facing errors
  actionable and structured at their source.

## Change quality

- Update documentation when behavior changes and add or supersede an ADR for a
  significant architectural decision.
- Keep tests meaningful and deterministic. Never call `os.Exit` outside the
  command entry point.
- Run formatting, tests, race tests, vet, lint when available, and builds before
  finishing. Inspect the diff for generated binaries, secrets, machine-specific
  paths, and accidental scope expansion.
- Do not commit generated binaries or release archives.
- Report assumptions, unavailable checks, incomplete work, and observed failures
  honestly. Never claim an unrun check succeeded.
