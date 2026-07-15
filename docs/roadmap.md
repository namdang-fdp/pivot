# Roadmap

Pivot is delivered as vertical slices. A slice is complete only after its user
path, failure cases, documentation, and safety behavior have been dogfooded on
real projects.

## Slice 0 — Foundation

- **User value:** a trustworthy repository, minimal `version` CLI, product
  contract, safety rules, and contributor/release foundations.
- **Main commands:** `pivot`, `pivot --help`, `pivot version`.
- **Exit criteria:** automated tests and checks pass; architecture, safety, ADRs,
  governance, CI, and release configuration agree.
- **Excluded:** all project configuration, discovery, state, inspection, and
  lifecycle behavior.

## Slice 1 — Registry, manifest, and doctor

**Implementation status:** Implemented. Maintainer dogfooding on real projects
remains required before the slice is considered complete for release.

- **User value:** explicitly register projects, validate strict manifests, list
  known projects, and diagnose requirements without starting anything.
- **Main commands:** `init`, `add`, `list`, `doctor`.
- **Exit criteria:** malformed input is rejected actionably; diagnosis is
  side-effect free; real projects can be configured and inspected.
- **Excluded:** starting, stopping, switching, persisted active state, and runtime
  resource ownership operations.

## Slice 2 — Single-project up and down

- **User value:** start one configured context, wait for readiness, and stop its
  proven resources while preserving data.
- **Main commands:** `up`, `down`.
- **Exit criteria:** operations are idempotent, Compose backends work on real
  projects, readiness is meaningful, and volumes survive shutdown.
- **Excluded:** cross-project switching, broad machine status, and rollback.

## Slice 3 — Ownership, reconciliation, and machine status

- **User value:** understand managed, unmanaged, partial, orphaned, and stale
  resources across the machine.
- **Main commands:** `status` and lifecycle status enhancements.
- **Exit criteria:** observation reconciles safely with records; ownership proof
  gates every destructive operation; drift scenarios are dogfooded.
- **Excluded:** automated cross-project transition and continuous monitoring.

## Slice 4 — Transactional switch

- **User value:** safely move from an active project to a ready target with
  preflight and best-effort rollback.
- **Main commands:** `switch`.
- **Exit criteria:** every transition phase and failure point is tested; source
  remains untouched on preflight failure; active state commits after readiness.
- **Excluded:** login startup, daemon coordination, and workspace automation.

## Slice 5 — Resume, shutdown, and workspace hooks

- **User value:** restore or stop the known active context and invoke explicit,
  bounded editor/workspace integrations.
- **Main commands:** `resume`, `shutdown`.
- **Exit criteria:** no surprise startup occurs; hooks call application services;
  flows are dogfooded across restarts.
- **Excluded:** terminal multiplexing, automatic login startup, and IDE plugins.

## Slice 6 — Logs and actionable failures

- **User value:** diagnose transition phases and runtime failures from structured,
  useful output.
- **Main commands:** lifecycle commands gain consistent diagnostic options.
- **Exit criteria:** errors name phase, cause, resulting state, and recovery;
  sensitive values are redacted; real failure scenarios are dogfooded.
- **Excluded:** hosted telemetry, remote log storage, and continuous monitoring.

## Slice 7 — Release hardening

- **User value:** a documented, reproducible, trustworthy v0.1 Linux release.
- **Main commands:** the complete v0.1 command set is hardened, not expanded.
- **Exit criteria:** compatibility, upgrades, artifacts, SBOMs, security review,
  and end-to-end acceptance scenarios pass on amd64 and arm64.
- **Excluded:** macOS/Windows guarantees, package registries, plugins, daemon,
  TUI, desktop, Kubernetes, and post-v0.1 product expansion.
