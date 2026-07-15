# Integration tests

This directory contains isolated, end-to-end CLI tests. Slice 1 tests create
temporary project roots and set `XDG_CONFIG_HOME` to a temporary directory, so
they never modify the developer's real Pivot registry. No live development
runtime is invoked. Future lifecycle integration tests must prove ownership and
data preservation before they invoke any destructive-capable adapter.
