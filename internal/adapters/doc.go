// Package adapters contains infrastructure implementations of application
// ports. Slice 1 adapters provide strict YAML manifests, an XDG registry,
// canonical path checks, and read-only command and TCP inspection. Future
// adapters may integrate external runtimes, lifecycle state, and workspaces.
// Filesystem and external process access belong at this boundary, not the core.
package adapters
