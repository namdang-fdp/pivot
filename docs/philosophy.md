# Product philosophy

These principles are permanent constraints, not preferences to revisit for each
feature.

## Orchestrate, do not replace

Pivot coordinates mature runtimes and project-owned commands. It does not absorb
their scheduling, build, package, or service-management responsibilities.

## Safety before magic

Convenience never justifies an action whose ownership or impact is uncertain.
No ownership proof means no destructive action.

## Explicit over implicit

Projects and behavior should be declared and reviewable. Repository
auto-detection may offer advice, but must not silently enroll or operate a
project.

## Reality is the source of truth

Persisted state expresses intent and history; observed machine state determines
what exists now. Pivot reconciles the two and reports disagreement.

## Idempotent lifecycle

Repeating a lifecycle command should converge on the requested state without
duplicating resources or turning ordinary drift into damage.

## No surprise startup

Observation, listing, diagnosis, and status operations must not start resources.
Startup requires a clear user request.

## Local-first and offline-first

Core operation must not require a hosted service, account, or network connection
beyond dependencies that the user's own project needs.

## Actionable errors

An error should name the failed phase, describe observed state, explain what was
left running, and suggest a safe next action. Raw runtime output alone is not a
diagnosis.

## Preserve developer data by default

Lifecycle operations stop compute while retaining volumes, images, source files,
and other durable data unless a future separately designed feature establishes
stronger consent and safeguards.

## Build vertical slices and dogfood them

Each slice must deliver end-to-end user value, include tests and documentation,
and be exercised on real projects. A slice is not complete because its internal
types compile.
