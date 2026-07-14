## Summary

Describe the change and its observable result.

## Motivation

Explain the user or contributor problem.

## Scope

Name the roadmap slice and anything deliberately excluded.

## Tests

List checks and scenarios run, including negative cases.

## Documentation

List updated docs or explain why behavior documentation is unaffected.

## Safety impact

Describe effects on ownership, processes, containers, commands, paths, state, and
developer data. Write "None" only after reviewing these areas.

## Relevant ADRs

Link existing or new decisions that constrain this change.

## Checklist

- [ ] The change supports a permanent Pivot product pillar.
- [ ] The change stays within the requested slice.
- [ ] Tests cover new behavior and safety-sensitive failures.
- [ ] Formatting, vet, tests, race tests, lint, and build pass locally, or unavailable checks are disclosed.
- [ ] Documentation matches user-visible behavior.
- [ ] Significant architecture changes include an ADR.
- [ ] No unmanaged resource can be modified without ownership proof.
- [ ] No generated binary, release archive, credential, or machine-specific file is committed.
