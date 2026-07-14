// Package core contains Pivot's domain concepts and invariants.
//
// The core is independent of presentation and infrastructure. It does not know
// about Cobra, Wails, YAML, Docker, Process Compose, the user's home directory,
// shell execution, or terminal output. Future domain behavior belongs here only
// when it can be expressed without those implementation details.
package core
