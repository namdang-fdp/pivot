# Test data

Shared fixtures belong here only when generated temporary fixtures would obscure
a test's intent. Slice 1 tests generate manifests and registries in temporary
directories to keep path and concurrency cases explicit. Any future fixture must
contain no credentials, host-specific paths, or real resource identifiers.
