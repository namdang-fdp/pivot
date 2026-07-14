# Architecture

Pivot is a Go modular monolith organized with ports and adapters. A single
process keeps deployment and debugging simple while package boundaries preserve
the option to add presentations or move coordination behind a service boundary
later.

```text
 CLI ───────┐
 TUI ───────┼──> Application services ──> Core
 Desktop ───┘              │                 │
                           └────> Ports <────┘
                                    │
                                Adapters
                  (runtimes, host, files, state, workspace)
```

Dependencies point inward. Presentation calls application services and renders
their results. Application services coordinate use cases using core concepts and
ports. Ports describe technology-neutral capabilities required by those use
cases. Adapters implement ports using external processes, Linux facilities, and
filesystems. The composition root wires concrete implementations.

## Layers

### Presentation

The CLI owns Cobra command definitions, input parsing, and terminal rendering.
Future TUI and desktop applications will call the same application services; the
desktop must never execute the `pivot` binary as its API. Presentation contains
no lifecycle decisions.

### Application services

Application services will coordinate lifecycle and transition use cases. They
decide sequencing and transaction boundaries but do not build Docker commands,
inspect host paths directly, or format terminal output.

### Core

The core will contain project state, ownership, transition, and safety concepts.
It must be deterministic and infrastructure-independent. It cannot depend on
Cobra or Wails because presentations are replaceable. It cannot depend on Docker
because runtime selection is an adapter concern. It cannot execute commands,
read `$HOME`, parse YAML, or print to the terminal.

### Ports and adapters

Ports are introduced with real use cases, not in anticipation of them. Adapters
will eventually cover Docker Compose, Process Compose, Linux inspection, atomic
JSON state, strict YAML configuration, and workspace integrations. External
process and filesystem effects remain at this boundary.

Diagnostic logging will use the standard `log/slog` package. The core should
emit structured results or events; presentation and configured observers decide
where they are rendered.

## Package policy

All implementation remains under `internal/` during v0.1. Pivot has no stable
extension contract and therefore no public `pkg/` tree. A public SDK requires a
versioning policy, demonstrated consumers, and an accepted ADR.

## Future service boundary

v0.1 commands reconcile state directly with no daemon. If real requirements such
as concurrent clients, a system tray, notifications, or continuous monitoring
emerge, application services and ports can be hosted behind a local protocol.
That migration would add an adapter and composition boundary; it must not move
business rules into transport handlers or make a daemon inevitable.
