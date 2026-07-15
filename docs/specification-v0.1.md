# Pivot v0.1 specification

This document defines the intended v0.1 behavior. Slice 1 currently implements
help, version reporting, strict manifest creation, project registration/listing,
and read-only doctor diagnostics. Lifecycle and transition requirements remain a
contract for later vertical slices, not a claim of implementation.

## Implemented commands

- `pivot init` creates explicit project configuration scaffolding.
- `pivot add` adds an explicitly selected project to the local registry.
- `pivot list` lists registered projects without starting them.
- `pivot doctor` validates configuration, tools, and conflicts without mutation.
- `pivot version` reports build metadata.

## Planned commands

- `pivot up` converges one project toward ready.
- `pivot down` stops only resources owned by one project and preserves data.
- `pivot status` reconciles recorded and observed machine state.
- `pivot switch` transactionally changes the active project.
- `pivot resume` restores the explicitly recorded active project on request.
- `pivot shutdown` stops the active project's owned resources.

The planned commands above are not available in Slice 1.

## Slice 1 persistence and manifest contract

Each project uses strict schema-version-1 YAML at `<project-root>/.pivot.yaml`.
Unknown fields, duplicate mapping keys, invalid project IDs, unsafe paths,
unsupported protocols, and unsupported versions fail deterministically. The
global project registry is versioned YAML at
`$XDG_CONFIG_HOME/pivot/projects.yaml`, falling back to
`~/.config/pivot/projects.yaml`. Registry updates use an inter-process lock and
atomic replacement. Slice 1 creates no runtime state or active-project record.

## State model

A project may be **stopped**, **starting**, **ready**, **stopping**, **partial**,
**failed**, or **unknown** based on reconciled observation. Transitional values
describe an operation in progress, not durable readiness. `unknown` means Pivot
lacks adequate evidence; it is never permission to act destructively.

The **active project** is the project whose complete context Pivot most recently
committed after successful readiness. The record is an intent/history fact and
must be reconciled with machine reality. At most one project is committed active,
but failures can leave resources from multiple projects present and must be
reported truthfully. No project becomes active merely because startup began.

**Managed resources** have verifiable project ownership through runtime namespace
and reconciled identity. **Unmanaged resources** lack that proof and are
read-only from Pivot's perspective, even when they conflict with a requested
transition.

## Functional requirements

- **FR-001 — Explicit registry:** Pivot shall operate only on explicitly
  registered projects and shall not silently enroll repositories.
- **FR-002 — Strict manifest:** Pivot shall validate the versioned project
  manifest strictly and report unknown or invalid fields actionably.
- **FR-003 — Side-effect-free diagnosis:** `list`, `doctor`, and read-only status
  operations shall not start or stop project resources.
- **FR-004 — Runtime delegation:** v0.1 shall delegate container lifecycle to the
  Docker Compose CLI and native process lifecycle to Process Compose.
- **FR-005 — Idempotent lifecycle:** repeated `up` and `down` requests shall
  converge without duplicating resources or treating already-achieved state as
  failure.
- **FR-006 — Readiness gate:** Pivot shall consider a project ready only after
  declared readiness conditions succeed within explicit bounds.
- **FR-007 — Ownership gate:** Pivot shall stop or signal only resources whose
  ownership by the selected project is proven.
- **FR-008 — Preserve data:** ordinary lifecycle commands shall preserve Docker
  volumes and other durable development data.
- **FR-009 — Reconciliation:** Pivot shall compare stored records with observed
  runtime and host state and expose partial, orphaned, stale, and unmanaged
  conditions.
- **FR-010 — Preflight:** `switch` shall validate the target and foreseeable
  conflicts before stopping the active source.
- **FR-011 — Transaction phases:** `switch` shall resolve, preflight, snapshot,
  stop, start, wait, commit, and on failure attempt rollback and reconciliation.
- **FR-012 — Active commit:** Pivot shall update the active project only after
  target readiness succeeds.
- **FR-013 — Truthful failure:** failed lifecycle operations shall report the
  observed resulting state and any failed rollback work.
- **FR-014 — Explicit resume:** Pivot shall not start the recorded project at
  login or during observation; `resume` requires a user request.
- **FR-015 — Structured diagnostics:** application and core behavior shall expose
  structured results or events that presentations can render.
- **FR-016 — Controlled process identity:** process operations shall verify more
  than a stored PID and shall account for PID reuse.
- **FR-017 — Configuration/state durability:** state writes shall use atomic JSON
  replacement with file locking and defensive validation.
- **FR-018 — Exit behavior:** command failures shall return non-zero status with
  actionable, non-duplicated errors.

## Non-functional requirements

- **NFR-001 — Platform:** v0.1 officially supports Linux on amd64 and arm64.
- **NFR-002 — Compatibility:** source builds require Go 1.25 or later and avoid
  CGO unless a separately accepted requirement proves it necessary.
- **NFR-003 — Safety:** absence or ambiguity of ownership evidence shall fail
  closed without destructive action.
- **NFR-004 — Local operation:** normal operation shall be local-first and usable
  offline except where project-owned dependencies require a network.
- **NFR-005 — Testability:** application behavior shall be testable without Cobra,
  process exit, or mandatory live runtimes; safety boundaries require integration
  coverage before release.
- **NFR-006 — Architecture:** dependencies shall point from presentations and
  adapters toward application ports and core, never from core to infrastructure.
- **NFR-007 — Observability:** diagnostics shall identify operation phase, relevant
  resource, underlying cause, and safe recovery guidance without leaking secrets.
- **NFR-008 — Responsiveness:** all waits shall be bounded, context-aware, and
  interruptible without falsifying state.
- **NFR-009 — Compatibility discipline:** persisted formats shall be versioned and
  invalid or newer formats shall fail clearly rather than be guessed.
- **NFR-010 — Least impact:** commands shall inspect and modify the smallest
  resource namespace necessary for the requested project.

## Acceptance scenarios

1. **Target preflight fails:** given project A is active and target B has invalid
   configuration or a known conflict, switching to B returns an actionable error;
   A remains untouched and recorded active.
2. **Successful switch:** given A is ready and B passes preflight, Pivot stops
   only A-owned compute, starts B, proves readiness, and then records B active.
3. **Target readiness fails:** after A is stopped and B starts but never becomes
   ready, Pivot attempts to stop B-owned resources and restore A, then reports
   every failure and reconciled state.
4. **Unmanaged conflict:** when an unrelated container or process occupies a
   required resource, Pivot reports the conflict and does not terminate it.
5. **Stale record:** when state names a PID or container that no longer matches
   ownership evidence, Pivot classifies the record stale and takes no destructive
   action against the observed replacement.
6. **Repeated lifecycle:** running `up` against a ready project or `down` against a
   stopped project succeeds without duplicate resources or data loss.
7. **Data preservation:** project shutdown and switching leave named and anonymous
   development volumes intact.
8. **No surprise startup:** listing, diagnosis, status, and machine restart do not
   automatically start any project.

## Definition of Done for v0.1

All planned commands meet their functional contracts; architecture and safety
tests cover success, interruption, drift, ambiguous ownership, and rollback;
Linux amd64 and arm64 artifacts are reproducible with checksums and SBOMs;
configuration and state compatibility are documented; security-sensitive paths
are reviewed; user and contributor documentation match behavior; CI and release
checks pass; and the complete workflow has been dogfooded on representative real
projects with both intended runtime backends.

## Explicit exclusions

v0.1 excludes a daemon, login startup, desktop UI, TUI, plugin system,
Kubernetes, Git worktree management, runtime installation, custom process
scheduling, automatic port allocation, secret management, AI repository
scanning, destructive cleanup commands, a public SDK, package-manager publishing,
hosted services, and support guarantees for non-Linux operating systems.
