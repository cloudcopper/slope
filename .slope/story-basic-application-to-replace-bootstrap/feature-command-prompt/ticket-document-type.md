# Type Document

For operating with the achetypes and tickets we need a type Document.
It shall represent property of .md files used for archetype and tickets:
- Filepath
- Metadata from frontmatter, if any
- Text

The type shall me in package internal/doc.
There shall be public constructor ```NewFromFile(fsys billy.Filesystem, filename string) (doc.Document, error)```. It shall load document from file or return an error. It shall support TOML/YAML frontmatter autodetection. It shallg return error if fronmatter is broken.

The Documant.Metadata shall be type of doc.Metadata map[string]string
Loading metadata shall work with flat K/V (TOML or only first level YAML),
and fail if there is complex/nested metadata (complex YAML).

The use of goldmark methods must be maximised.
Consder to use https://github.com/abhinav/goldmark-frontmatter for TOML/YAML operations.

The doc.Document shall have method ```ID() string``` which returns document ID according to spec (from metadata, and then from filename if missing).
The doc.Document shall have method ```Archetype() string``` which returns document archetype according to spec (from metadata, and then from filename if missing).

The golang related skills MUST BE USED

# Testing

Must have tests according to testing skills.
Must have 100% test coverage (may requires to add target into Makefile).
Test must use internal/utils/testutils/FsBuilder with billy memfs (see in that package how to use it). No real disk usage allowed for test.
Must have no dead code.

# Final checks

Complete when:
- [x] ```make build``` produce ./slope
- [x] ```make test``` has no issues
- [x] ```make cover``` works and the internal/fs has 100% coverage
