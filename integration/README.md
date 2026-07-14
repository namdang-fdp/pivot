# Integration tests

This directory will contain Linux integration tests that exercise real adapter
boundaries in isolated environments. Slice 0 has no runtime adapters to test;
the minimal CLI is covered by unit tests. Future tests must prove ownership and
data-preservation behavior before they invoke destructive operations.
