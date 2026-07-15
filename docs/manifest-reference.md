# Manifest reference

Pivot Slice 1 reads one project manifest at `<project-root>/.pivot.yaml`. The
format is strict YAML: unknown fields, duplicate mapping keys, multiple YAML
documents, unsupported versions, and invalid values are rejected. The matching
editor schema is [`schemas/pivot.schema.json`](../schemas/pivot.schema.json).

## Complete Slice 1 example

```yaml
version: 1

project:
  id: centeros
  name: CenterOS

requirements:
  commands:
    - docker
    - process-compose
    - mise
  files:
    - .env

compose:
  files:
    - ops/compose.dev.yml
  envFiles:
    - .env

ports:
  - name: postgres
    port: 5432
    protocol: tcp
    required: true
  - name: gateway
    port: 8080
    protocol: tcp
    required: true
```

## Top-level fields

`version`, `project`, `requirements`, and `ports` are required. `compose` is
optional. No other top-level fields are accepted.

### `version`

Required integer. Slice 1 accepts only `1`; missing or newer versions fail
without guessing compatibility.

### `project`

Required object with exactly these fields:

- `id` is required and must match `^[a-z0-9]+(?:-[a-z0-9]+)*$`. Loading never
  normalizes an invalid ID. Only `pivot init` derives a slug.
- `name` is required and must contain a non-whitespace character. It is the
  human-readable name stored in the registry.

### `requirements`

Required object with `commands` and `files` arrays, which may be empty.

- `commands` contains non-empty, unique executable names resolved through
  `PATH` by `pivot doctor`.
- `files` contains non-empty, unique paths to files required beneath the project
  root. Doctor checks existence but never reads or prints file contents.

### `compose`

Optional object. `files` lists unique Compose files and `envFiles` lists unique
Compose environment files. Both arrays may be omitted or empty. Doctor checks
the declared files, the `docker` command, and `docker compose version` whenever
at least one Compose file is declared. Slice 1 does not start or stop Compose.

### `ports`

Required array, which may be empty. Each entry has exactly:

- `name`: non-empty and unique within the manifest;
- `port`: integer from 1 through 65535;
- `protocol`: `tcp` (the only Slice 1 protocol);
- `required`: boolean indicating whether doctor must prove availability.

An occupied required port is a failed diagnostic and is always considered
unmanaged in Slice 1. Pivot may enrich the report with a PID and command but
never signals or terminates the owner.

## Path safety

All values in `requirements.files`, `compose.files`, and `compose.envFiles` are
project-relative. Empty paths, absolute paths, normalized duplicates, and paths
whose lexical form traverses above the root are invalid. Pivot also evaluates
each existing path component; an existing symlink that resolves outside the
canonical project root is rejected. Missing leaf paths remain valid manifest
declarations so doctor can report them as missing.

## Minimal generated manifest

`pivot init` creates only:

```yaml
version: 1
project:
  id: centeros
  name: CenterOS
requirements:
  commands: []
  files: []
ports: []
```

It does not inspect source code, infer tools or ports, create a local override,
or register the project.
