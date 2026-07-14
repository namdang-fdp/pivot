# Safety model

Pivot coordinates processes and containers on a developer's machine. A mistaken
destructive action can lose work or data, so safety is a product invariant:

> **No ownership proof, no destructive action.**

## Resource classifications

- **Managed:** observed resource with current, verifiable evidence that a known
  Pivot project owns it.
- **Unmanaged:** resource outside Pivot's ownership boundary. Pivot may report it
  but must not stop or modify it.
- **Partial:** only part of a project's declared context is present or healthy.
- **Orphaned:** ownership can be tied to a known project, but the expected parent
  operation or current state record is absent.
- **Stale:** persisted state claims a resource or condition that observation no
  longer confirms.

Classification is evidence, not permission by itself. Ambiguous resources are
treated as unmanaged until ownership is proven.

## Ownership evidence

Docker resources will be scoped through an explicit Docker Compose project
namespace derived from Pivot's project identity. Compose labels and inspection
must agree with the intended project before a stop operation. Pivot never treats
“currently running Docker container” as equivalent to “Pivot-owned container.”

The future Process Compose adapter will use a per-project namespace and socket or
equivalent control endpoint. A stored PID alone is insufficient: process
identity, start metadata, namespace, and observed command provenance must be
reconciled because operating systems can reuse PIDs.

State records support reconciliation but never override contradictory machine
evidence. Missing, stale, or corrupt state reduces permission to act; it does not
broaden it.

## Forbidden operations

Pivot must not use `pkill` or `killall`, terminate by generic executable name,
stop all Docker containers, or invoke global volume/image pruning. Those actions
cannot establish project ownership and can affect unrelated workloads. v0.1 has
no `--force-kill`: a flag cannot turn uncertain identity into proof, and it would
make unsafe behavior part of the public contract.

## Data preservation

Normal lifecycle transitions stop compute but preserve Docker volumes, images,
source trees, caches, databases, and other durable developer data. Destructive
cleanup commands are excluded from v0.1. Runtime command construction and tests
must make the absence of volume deletion explicit.

## Reconciliation and truthful state

Pivot compares persisted intent with observed resources before deciding what is
active, partial, orphaned, stale, or unmanaged. It must never report a target as
active solely because startup was requested. After any failure, status reflects
what observation can prove, including mixed or partial states.

## Transaction safety

Before a future switch changes the active source, Pivot preflights target
configuration, dependencies, conflicts, and readiness definitions. A preflight
failure leaves the source untouched. The transition snapshots sufficient source
state, stops only source-owned resources, starts the target, waits for readiness,
and commits the active target only after success.

If a later phase fails, rollback is best-effort rather than fictional atomicity:
Pivot attempts to stop target-owned resources it started and restore the source,
then reconciles reality. Rollback errors are reported alongside the initiating
error. Pivot never falsifies state to make the transaction appear successful.
