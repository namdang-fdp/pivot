# ADR-0001: Product boundary

## Status

Accepted

## Context

Developers moving between local projects need coordination across containers,
native processes, ports, readiness, and recorded context. Adjacent developer
tools already solve builds, runtime installation, scheduling, terminal sessions,
cluster operation, and source-tree organization. Absorbing those concerns would
weaken Pivot's safety model and make its purpose incoherent.

## Decision

Pivot is a machine-level local project context orchestrator. It observes machine
reality, coordinates project lifecycle through existing tools, and performs safe
context transitions. Every feature must support machine awareness, project
lifecycle, or safe context transition.

Pivot must not become a Make replacement, runtime installer, process manager,
terminal multiplexer, Kubernetes dashboard, or Git worktree manager.

## Consequences

Project-owned build and task definitions remain authoritative. Pivot integrates
with rather than reimplements runtimes. Some seemingly convenient features are
rejected when they cross the boundary, keeping the core smaller and safety work
focused.

## Alternatives considered

- A universal local development platform was rejected as too broad and likely to
  duplicate mature tools.
- A task runner was rejected because command invocation alone cannot establish
  ownership, readiness, or transactional context.
- A Docker-only manager was rejected because a project context also includes
  native processes and host constraints.
