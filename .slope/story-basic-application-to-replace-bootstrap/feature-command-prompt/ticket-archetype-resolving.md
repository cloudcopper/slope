# Archetype resolving

The archetype resolving is needed before prompt command can work.
Per IDEA.md §3.3, archetypes are resolved from:
- .slope/archetype/<name>.md (project-local)
- ~/.config/slope/archetype/<name>.md (user-global)

It shall be developed in package internal/archetype
Use golang skills for developing and test

The package should have method ```func Find(fsys billy.Filesystem, root string, archetype string) (document.Document, error)```
The method shall lookup document as per IDEA.md,
load it,
and return (or return error)

Error shall be done similar to internal/document.

# Testing

Must have tests according to testing skills.
Must have 100% test coverage.
Test must use internal/utils/testutils/FsBuilder with billy memfs (see in that package how to use it). No real disk usage allowed for test.
Must have no dead code.

# Final checks

Complete when:
- [ ] ```make build``` produce ./slope
- [ ] ```make test``` has no issues
- [ ] ```make cover``` works and the internal/fs has 100% coverage



