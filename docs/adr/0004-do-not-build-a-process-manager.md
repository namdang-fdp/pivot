# ADR-0004: Do not build a process manager

## Status

Accepted

## Context

Native development processes need grouped startup, dependency ordering, output,
restart policy, and signal handling. Building those capabilities would turn
Pivot into a scheduler and divert effort from ownership and safe transitions.

## Decision

Process Compose is the intended v0.1 backend for native processes. Pivot owns
project-level lifecycle semantics: when a context should start or stop, how
ownership is proven, how readiness affects state, and how failures participate in
a transition. Process Compose owns scheduling and supervision internals.

## Consequences

Pivot requires an adapter and explicit per-project Process Compose namespaces or
control endpoints. Users retain Process Compose configuration and behavior.
Pivot will surface backend failures actionably but will not recreate restart,
dependency, logging, or signal-management machinery.

## Alternatives considered

- A custom supervisor was rejected as a second product with substantial safety
  and portability burden.
- Raw background shell processes were rejected because reliable grouping,
  identity, output, and teardown would recreate a supervisor poorly.
- Systemd user services were rejected as too persistent and host-coupled for
  ephemeral project development contexts.
