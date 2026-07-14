# ADR-0003: No daemon in v0.1

## Status

Accepted

## Context

A daemon could cache state or watch the machine, but it introduces startup,
upgrade, authentication, concurrency, crash recovery, and stale-observation
problems before v0.1 has proven its lifecycle model. Background behavior also
conflicts with the no-surprise-startup principle.

## Decision

v0.1 has no daemon. Each command reads durable state, observes relevant machine
reality, reconciles them, performs its bounded operation, and writes state
directly. There is no background polling and nothing starts at login.

A daemon may be proposed only when concrete requirements such as a system tray,
concurrent clients, notifications, or continuous monitoring justify it.

## Consequences

Commands own synchronization and cannot rely on in-memory truth between runs.
File locking and atomic state become important. Continuous updates and unsolicited
notifications are unavailable in v0.1, while operation remains transparent and
offline-friendly.

## Alternatives considered

- A daemon from the first release was rejected as operational complexity without
  demonstrated user value.
- Optional background polling was rejected because “optional” still creates two
  consistency models and surprise behavior.
- A login service was rejected because Pivot never starts projects implicitly.
