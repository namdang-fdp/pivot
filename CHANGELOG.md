# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project intends to follow [Semantic Versioning](https://semver.org/)
after a public versioned release exists.

## [Unreleased]

### Added

- Slice 1 strict `.pivot.yaml` schema and manifest validation, including path
  traversal and symlink escape protection.
- XDG project registry with deterministic YAML, atomic writes, file locking, and
  duplicate ID/path protection.
- `pivot init`, `pivot add`, `pivot list`, and read-only `pivot doctor` commands
  with stable JSON output for list and doctor.
- Read-only host command, Compose capability, required-file, and TCP port
  diagnostics with best-effort Linux owner enrichment.
- Slice 1 command/manifest documentation, JSON Schema, and isolated integration
  tests.
- Slice 0 repository foundation for the Pivot modular Go monolith.
- Minimal Cobra CLI with help and human-readable or JSON version output.
- Product, architecture, safety, specification, roadmap, and ADR documentation.
- Open-source governance, contributor tooling, CI, and release configuration.
- Unit tests for build information and CLI presentation behavior.
