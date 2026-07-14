# ADR-0005: Resource ownership

## Status

Accepted

## Context

Developer machines run containers and processes unrelated to the current Pivot
project. Names, ports, and PIDs can collide or be reused. Persisted state can
become stale. A context orchestrator that guesses ownership can destroy work.

## Decision

> No ownership proof, no destructive action.

Docker Compose projects use explicit Pivot-derived project namespaces, verified
against observed Compose labels before modification. Future Process Compose
integration uses per-project namespaces and sockets or equivalent endpoints,
with process identity reconciled beyond a PID. State records contribute evidence
but cannot overrule contradictory observation. Resources without adequate proof
are unmanaged and may be reported but not stopped.

## Consequences

Some operations will refuse to proceed and require the user to resolve ambiguous
or unmanaged conflicts. Adapters must expose evidence, not only booleans. Tests
must cover stale state, PID reuse, namespace collisions, partial resources, and
unmanaged workloads. Broad stop and prune operations are prohibited.

## Alternatives considered

- Trusting recorded PIDs or container names was rejected because identifiers are
  reusable and state can be stale.
- Treating port conflicts as ownership was rejected because a port proves use,
  not project identity.
- A `--force-kill` escape hatch was rejected for v0.1 because consent does not
  resolve ambiguous identity and would normalize unsafe recovery.
