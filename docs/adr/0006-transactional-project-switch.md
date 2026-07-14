# ADR-0006: Transactional project switch

## Status

Accepted

## Context

Switching projects can fail after the source stops but before the target is
ready. A linear script may leave conflicting resources, lose the active context,
or record success prematurely. Perfect atomicity is impossible across external
runtimes, so the design needs explicit phases and truthful recovery.

## Decision

A future switch executes these phases:

```text
resolve
preflight target
snapshot source
stop source
start target
wait for readiness
commit target active
rollback source on failure
reconcile truthful state
```

Preflight is read-only with respect to the source. If it fails, the active source
project is unaffected. The target becomes active only after readiness succeeds.
Failures after mutation trigger best-effort rollback using verified ownership,
followed by observation and truthful reconciliation.

## Consequences

Application services must make phases and compensation explicit. State commits
are late and durable. Errors can include both the initiating failure and rollback
failures. Tests must inject failure and interruption at every phase and prove that
unmanaged resources and data remain untouched.

## Alternatives considered

- Stop-then-start scripting was rejected because it lacks preflight, commit, and
  recovery semantics.
- Recording the target active before readiness was rejected because intent would
  be presented as fact.
- Claiming fully atomic switching was rejected because external processes and
  containers cannot participate in a single database transaction.
