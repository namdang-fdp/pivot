# Schemas

[`pivot.schema.json`](pivot.schema.json) is the editor-facing JSON Schema for
the strict Slice 1 `.pivot.yaml` manifest. The Go decoder and validator remain
authoritative; schema changes must ship with matching implementation tests and
manifest-reference documentation.
