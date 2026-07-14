# ADR-0002: Go and ports/adapters

## Status

Accepted

## Context

Pivot needs a portable single binary, reliable concurrency and process control,
fast startup, clear errors, and maintainable boundaries between safety rules and
host integrations. CLI, future TUI, desktop, daemon, and IDE presentations must
share behavior without one framework defining the core.

## Decision

The core is implemented in Go 1.25 or later as a modular monolith using ports and
adapters. Presentations call application services; application services depend
on core concepts and technology-neutral ports; infrastructure adapters implement
those ports. All implementation remains internal until a public API is designed.

Go is selected for its simple deployment, strong standard library, mature CLI
and systems ecosystem, approachable maintenance, and suitable process and
concurrency primitives. Desktop framework selection cannot control domain or
application design. Ports and adapters preserve options for CLI, TUI, desktop,
daemon, and IDE clients.

## Consequences

The core cannot import Cobra, Wails, Docker clients, or serialization libraries.
Adapters may be replaced and presentations remain peers. The repository accepts
some interface wiring, but interfaces are introduced only with a current use
case. v0.1 avoids CGO unless a proven requirement changes the decision.

## Alternatives considered

- Rust offers excellent control and safety but adds language and build complexity
  that is not justified for the initial contributor and integration profile.
- TypeScript aligns with some desktop frameworks but makes Node distribution or
  bundling a core concern and offers weaker boundaries for low-level host work.
- Choosing the desktop framework first was rejected because presentation
  lifetime and constraints should not dictate lifecycle semantics.
