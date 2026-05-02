+++
status = "done"
+++
# Document walk

The package internal/document need Walk functionality.

Expected function is
```func Walk(fsys billy.Filesystem, root string, fn WalkFunc) error```
where
```type WalkFunc func(fsys billy.Filesystem, doc Document, err error) error```
working similarly to os.Walk,
except it load document and pass it to callbak.

It shall try to load only 1) .md files (except README.md), then 2) README.md from directories (promoted tickets)
and in fact shall scan one level of tickets only.

# Skills

Must use skill ```golang-dev``` and mentioned there

# Testing

Must have tests according to testing skills.
Must have 100% test coverage.
Test must use internal/utils/testutils/FsBuilder with billy memfs (see in that package how to use it). No real disk usage allowed for test.
Must have no dead code.

# Final checks

Complete when:
- [x] ```make build``` produce ./slope
- [x] ```make test``` has no issues
- [x] ```make cover``` works and the internal/fs has 100% coverage

# Final report

One-shot was corrected with HITL (remove unnecesery if and small simplification of asserts in tests)
