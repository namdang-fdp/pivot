# Command reference

Pivot is pre-alpha. Slice 1 implements only the commands documented here; the
lifecycle commands in the roadmap are planned and are not usable yet.

## `pivot init [path]`

Creates a minimal `.pivot.yaml` in an existing directory. The path defaults to
the current directory. `--id` and `--name` override values derived from the
directory basename. Init refuses non-directories and existing manifests, writes
atomically, validates its own output, and does not register the project.

```bash
pivot init ~/Projects/centeros --id centeros --name CenterOS
pivot add ~/Projects/centeros
```

## `pivot add [path]`

Strictly validates `<path>/.pivot.yaml` and registers its canonical absolute
project path. The path defaults to the current directory. Re-adding the same ID
and path succeeds idempotently. Reusing an ID for another path, or a path for
another ID, fails without changing the registry.

The registry is `$XDG_CONFIG_HOME/pivot/projects.yaml`, falling back to
`~/.config/pivot/projects.yaml`. Tests and experiments can be isolated with:

```bash
XDG_CONFIG_HOME=<temp-directory> pivot list
```

Registry writes are locked, deterministic, and atomically replaced.

## `pivot list`

Lists registered projects by ID. `AVAILABLE` means only that the project
directory and `.pivot.yaml` currently exist; it is not runtime health.

```bash
pivot list
pivot list --json
```

An empty registry is successful and explains how to register a project.

## `pivot doctor [project]`

Performs read-only checks for the strict manifest, registry consistency,
required commands and files, Compose declarations/tooling, and required TCP
port availability.

```bash
pivot doctor centeros
pivot doctor centeros --json
```

Without an ID, doctor selects a registered project containing the current
directory. If none matches, it may inspect a `.pivot.yaml` found in the current
directory or an ancestor without registering it. Failed checks produce a
non-zero exit; warnings do not. Occupied ports are unmanaged and never
terminated. Doctor never repairs configuration, installs tools, starts services,
or rewrites project files.

## `pivot version`

Prints build metadata. `pivot version --json` provides stable field names for
automation.
