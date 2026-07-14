# Product definition

## Target persona

Pivot serves developers who regularly work with five to fifteen local projects
on one Linux machine. Their projects commonly mix Docker Compose infrastructure,
native application processes, ports, readiness checks, and editor workspaces.
They value fast context changes but cannot trade away local data or control of
unrelated workloads.

## Problem and user pain

Project startup knowledge is fragmented across documentation, shell history,
task files, and memory. Developers must discover what is running, stop conflicts,
sequence dependencies, verify readiness, and remember which project was active.
A partial transition leaves the machine ambiguous. Broad cleanup commands can
destroy data or terminate unrelated work.

## Product promise

Pivot safely transitions a developer's machine between complete local
development contexts. It observes real machine state, applies explicit project
intent through existing runtimes, and records an active project only after the
target is ready.

The north-star experience is:

```bash
pivot switch smartpark
```

Pivot will resolve the source and target, preflight the target without disturbing
the source, stop only source-owned resources, start the target, wait for
readiness, commit the target as active, and attempt a safe rollback on failure.

## Product pillars

1. **Machine awareness** — distinguish observed state from recorded intent and
   reconcile differences explicitly.
2. **Project lifecycle** — coordinate idempotent startup, readiness, and shutdown
   while delegating runtime internals.
3. **Safe context transition** — protect ownership and data across a transaction,
   with truthful failure state.

## Product boundary

Pivot is a machine-level context orchestrator. It composes existing tools; it
does not replace Make, install runtimes, schedule processes, manage terminal
sessions, operate Kubernetes, or manage Git worktrees. v0.1 is local-first,
Linux-only, and command-driven. Desktop, TUI, daemon, plugins, secrets, automatic
port allocation, and AI scanning are outside the release boundary.

## Why this is not merely a task runner

A task runner launches declared commands. Pivot reasons about a transition:
current machine reality, resource ownership, preconditions, readiness, state
commit, and recovery. Existing task recipes may remain useful inside projects;
Pivot operates at the context boundary around them.

## v0.1 success criteria

v0.1 succeeds when real projects can be registered explicitly, diagnosed,
started, stopped, switched, resumed, and shut down on Linux; managed resources
are distinguished from unmanaged ones; readiness gates active-state commits;
failures produce actionable explanations and truthful status; volumes and other
developer data are preserved by default; no destructive action occurs without
ownership proof; and the complete workflow has been dogfooded on representative
Docker Compose and Process Compose projects.
