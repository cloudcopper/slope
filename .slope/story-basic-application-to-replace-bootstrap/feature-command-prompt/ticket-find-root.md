+++
status = "done"
+++
# Find root functionality

We need a method in package internal/fs to find named root walking up.
It must use billy.Filesystem instead of golang stdlib fs,
so we can test it with inmemory filesystem (see internal/utils/testutil)

See similar functinality in bootstrap.go

Should be in dedicated file find_root.go
Should lookup for search (file or directory) from cwd up till "/".
Should return only stdlib errors
No windows support needed. Only unix atm.
Must be implemented according to golang skills.

```func FindRoot(fsys billy.Filesystem, cwd string, search path) (string, error)```

# Testing

Test file find_root_test.go
Must have tests according to testing skills.
Must have 100% test coverage (may requires to add target into Makefile).
Test must use internal/utils/testutils/FsBuilder with billy memfs (see in that package how to use it). No real disk usage allowed for test.

# Final checks

Complete when:
- [x] ```make build``` produce ./slope
- [x] ```make test``` has no issues
- [x] ```make cover``` works and the internal/fs has 100% coverage

