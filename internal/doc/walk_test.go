package doc_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/egorse/slope/internal/doc"
	"github.com/egorse/slope/internal/utils/testutil"
)

func TestWalk(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setup      func(*testutil.FsBuilder)
		root       string
		expectDocs []string // expected document IDs in order
		wantErrMsg string   // expected error message from Walk itself (not callback)
	}{
		{
			name:       "empty directory",
			setup:      func(b *testutil.FsBuilder) { b.Dir("/project/.slope") },
			root:       "/project/.slope",
			expectDocs: nil,
		},
		{
			name: "single .md file",
			setup: func(b *testutil.FsBuilder) {
				b.Dir("/project/.slope")
				b.File("/project/.slope/story-aaa.md", "# Hello\n")
			},
			root:       "/project/.slope",
			expectDocs: []string{"story-aaa"},
		},
		{
			name: "multiple .md files",
			setup: func(b *testutil.FsBuilder) {
				b.Dir("/project/.slope")
				b.File("/project/.slope/story-aaa.md", "# A\n")
				b.File("/project/.slope/feature-bbb.md", "# B\n")
				b.File("/project/.slope/todo-ccc.md", "# C\n")
			},
			root: "/project/.slope",
			expectDocs: []string{
				"feature-bbb",
				"story-aaa",
				"todo-ccc",
			},
		},
		{
			name: "directory with README.md",
			setup: func(b *testutil.FsBuilder) {
				b.Dir("/project/.slope")
				b.Dir("/project/.slope/story-abc")
				b.File("/project/.slope/story-abc/README.md", "---\n_id: story-abc\n---\n# ABC\n")
			},
			root:       "/project/.slope",
			expectDocs: []string{"story-abc"},
		},
		{
			name: "mixed .md files and directories with README.md",
			setup: func(b *testutil.FsBuilder) {
				b.Dir("/project/.slope")
				b.File("/project/.slope/story-aaa.md", "# A\n")
				b.Dir("/project/.slope/story-bbb")
				b.File("/project/.slope/story-bbb/README.md", "---\n_id: story-bbb\n---\n# B\n")
				b.File("/project/.slope/todo-ccc.md", "# C\n")
			},
			root: "/project/.slope",
			expectDocs: []string{
				"story-aaa",
				"story-bbb",
				"todo-ccc",
			},
		},
		{
			name: "README.md at root is skipped",
			setup: func(b *testutil.FsBuilder) {
				b.Dir("/project/.slope")
				b.File("/project/.slope/README.md", "# Root\n")
				b.File("/project/.slope/story-aaa.md", "# A\n")
			},
			root:       "/project/.slope",
			expectDocs: []string{"story-aaa"},
		},
		{
			name: "non-.md files are skipped",
			setup: func(b *testutil.FsBuilder) {
				b.Dir("/project/.slope")
				b.File("/project/.slope/story-aaa.md", "# A\n")
				b.File("/project/.slope/notes.txt", "notes\n")
				b.File("/project/.slope/diagram.png", "")
			},
			root:       "/project/.slope",
			expectDocs: []string{"story-aaa"},
		},
		{
			name: "directory without README.md is skipped",
			setup: func(b *testutil.FsBuilder) {
				b.Dir("/project/.slope")
				b.Dir("/project/.slope/empty-dir")
				b.File("/project/.slope/story-aaa.md", "# A\n")
			},
			root:       "/project/.slope",
			expectDocs: []string{"story-aaa"},
		},
		{
			name: "directory with non-README.md files is skipped",
			setup: func(b *testutil.FsBuilder) {
				b.Dir("/project/.slope")
				b.Dir("/project/.slope/story-bbb")
				b.File("/project/.slope/story-bbb/other.md", "# Other\n")
				b.File("/project/.slope/story-aaa.md", "# A\n")
			},
			root:       "/project/.slope",
			expectDocs: []string{"story-aaa"},
		},
		{
			name:       "root not found returns error",
			setup:      func(b *testutil.FsBuilder) {},
			root:       "/project/.slope",
			wantErrMsg: `doc: walk directory "/project/.slope": directory not found`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fsys := memfs.New()
			builder := testutil.NewFsBuilder(fsys)
			tt.setup(builder)

			must := require.New(t)
			is := assert.New(t)

			var docs []doc.Document
			err := doc.Walk(fsys, tt.root, func(fsys billy.Filesystem, d doc.Document, err error) error {
				if err != nil {
					return err
				}
				docs = append(docs, d)
				return nil
			})

			if tt.wantErrMsg != "" {
				must.Error(err)
				is.EqualError(err, tt.wantErrMsg)
				is.Empty(docs)
				return
			}

			must.NoError(err)
			is.Len(docs, len(tt.expectDocs))
			for i, id := range tt.expectDocs {
				is.Equal(id, docs[i].ID())
			}
		})
	}
}

func TestWalk_CallbackErrorStopsWalk(t *testing.T) {
	t.Parallel()

	fsys := memfs.New()
	testutil.NewFsBuilder(fsys).
		Dir("/project/.slope").
		File("/project/.slope/story-aaa.md", "# A\n").
		File("/project/.slope/feature-bbb.md", "# B\n").
		File("/project/.slope/todo-ccc.md", "# C\n")

	must := require.New(t)
	is := assert.New(t)

	var docs []doc.Document
	sentinel := errors.New("stop walking")
	err := doc.Walk(fsys, "/project/.slope", func(fsys billy.Filesystem, d doc.Document, err error) error {
		if err != nil {
			return err
		}
		docs = append(docs, d)
		if d.ID() == "feature-bbb" {
			return sentinel
		}
		return nil
	})

	must.Error(err)
	is.ErrorIs(err, sentinel)
	is.Len(docs, 1)
	is.Equal("feature-bbb", docs[0].ID())
}

func TestWalk_CallbackReceivesLoadError(t *testing.T) {
	t.Parallel()

	fsys := memfs.New()
	testutil.NewFsBuilder(fsys).
		Dir("/project/.slope").
		File("/project/.slope/story-aaa.md", "# A\n").
		File("/project/.slope/story-bad.md", "---\ntitle: Orphaned\nbody text\n")

	must := require.New(t)
	is := assert.New(t)

	var docs []doc.Document
	var loadErr error
	err := doc.Walk(fsys, "/project/.slope", func(fsys billy.Filesystem, d doc.Document, err error) error {
		if err != nil {
			loadErr = err
			t.Logf("load error: %v (type: %T)", err, err)
			return nil // don't stop, just record
		}
		docs = append(docs, d)
		t.Logf("loaded doc: %s", d.ID())
		return nil
	})

	must.NoError(err)
	is.Len(docs, 1)
	is.Equal("story-aaa", docs[0].ID())
	must.NotNil(loadErr)
	var target *doc.ErrFrontmatter
	is.True(errors.As(loadErr, &target))
	is.Equal("/project/.slope/story-bad.md", target.Filepath)
}

func TestWalk_DirectoryWithCorruptREADME(t *testing.T) {
	t.Parallel()

	fsys := memfs.New()
	testutil.NewFsBuilder(fsys).
		Dir("/project/.slope").
		Dir("/project/.slope/story-bad").
		File("/project/.slope/story-bad/README.md", "---\ntitle: Orphaned\nbody text\n").
		File("/project/.slope/story-good.md", "# Good\n")

	must := require.New(t)
	is := assert.New(t)

	var docs []doc.Document
	var loadErr error
	err := doc.Walk(fsys, "/project/.slope", func(fsys billy.Filesystem, d doc.Document, err error) error {
		if err != nil {
			loadErr = err
			return nil
		}
		docs = append(docs, d)
		return nil
	})

	must.NoError(err)
	is.Len(docs, 1)
	is.Equal("story-good", docs[0].ID())
	must.NotNil(loadErr)
	var target *doc.ErrFrontmatter
	is.True(errors.As(loadErr, &target))
	is.Equal("/project/.slope/story-bad/README.md", target.Filepath)
}

func TestWalk_WalkFuncSignature(t *testing.T) {
	t.Parallel()

	// Verify that WalkFunc is usable as a standalone function type.
	fsys := memfs.New()
	testutil.NewFsBuilder(fsys).
		Dir("/project/.slope").
		File("/project/.slope/story-aaa.md", "# A\n")

	must := require.New(t)

	var count int
	var fn doc.WalkFunc = func(fsys billy.Filesystem, d doc.Document, err error) error {
		count++
		return nil
	}

	must.NoError(doc.Walk(fsys, "/project/.slope", fn))
	is := assert.New(t)
	is.Equal(1, count)
}

func TestWalk_MultipleDirectoriesWithREADME(t *testing.T) {
	t.Parallel()

	fsys := memfs.New()
	testutil.NewFsBuilder(fsys).
		Dir("/project/.slope").
		Dir("/project/.slope/story-alpha").
		File("/project/.slope/story-alpha/README.md", "---\n_id: story-alpha\n---\n# Alpha\n").
		Dir("/project/.slope/story-beta").
		File("/project/.slope/story-beta/README.md", "---\n_id: story-beta\n---\n# Beta\n").
		File("/project/.slope/todo-gamma.md", "# Gamma\n")

	must := require.New(t)
	is := assert.New(t)

	var ids []string
	err := doc.Walk(fsys, "/project/.slope", func(fsys billy.Filesystem, d doc.Document, err error) error {
		if err != nil {
			return err
		}
		ids = append(ids, d.ID())
		return nil
	})

	must.NoError(err)
	is.Equal([]string{"story-alpha", "story-beta", "todo-gamma"}, ids)
}

func TestWalk_DotFilesAreSkipped(t *testing.T) {
	t.Parallel()

	fsys := memfs.New()
	testutil.NewFsBuilder(fsys).
		Dir("/project/.slope").
		File("/project/.slope/.hidden.md", "# Hidden\n").
		File("/project/.slope/story-aaa.md", "# A\n")

	must := require.New(t)
	is := assert.New(t)

	var ids []string
	err := doc.Walk(fsys, "/project/.slope", func(fsys billy.Filesystem, d doc.Document, err error) error {
		if err != nil {
			return err
		}
		ids = append(ids, d.ID())
		return nil
	})

	must.NoError(err)
	is.Equal([]string{"story-aaa"}, ids)
}

func TestWalk_RootWithTrailingSlash(t *testing.T) {
	t.Parallel()

	fsys := memfs.New()
	testutil.NewFsBuilder(fsys).
		Dir("/project/.slope").
		File("/project/.slope/story-aaa.md", "# A\n")

	must := require.New(t)
	is := assert.New(t)

	var ids []string
	err := doc.Walk(fsys, "/project/.slope/", func(fsys billy.Filesystem, d doc.Document, err error) error {
		if err != nil {
			return err
		}
		ids = append(ids, d.ID())
		return nil
	})

	must.NoError(err)
	is.Equal([]string{"story-aaa"}, ids)
}

func TestWalk_RootAsDot(t *testing.T) {
	t.Parallel()

	fsys := memfs.New()
	testutil.NewFsBuilder(fsys).
		Dir("/project/.slope").
		File("/project/.slope/story-aaa.md", "# A\n")

	must := require.New(t)
	is := assert.New(t)

	var ids []string
	err := doc.Walk(fsys, "/project/.slope", func(fsys billy.Filesystem, d doc.Document, err error) error {
		if err != nil {
			return err
		}
		ids = append(ids, d.ID())
		return nil
	})

	must.NoError(err)
	is.Equal([]string{"story-aaa"}, ids)
}

func TestWalk_SkipHiddenDirectories(t *testing.T) {
	t.Parallel()

	fsys := memfs.New()
	testutil.NewFsBuilder(fsys).
		Dir("/project/.slope").
		Dir("/project/.slope/.hidden-dir").
		File("/project/.slope/.hidden-dir/README.md", "---\n_id: hidden\n---\n").
		File("/project/.slope/story-aaa.md", "# A\n")

	must := require.New(t)
	is := assert.New(t)

	var ids []string
	err := doc.Walk(fsys, "/project/.slope", func(fsys billy.Filesystem, d doc.Document, err error) error {
		if err != nil {
			return err
		}
		ids = append(ids, d.ID())
		return nil
	})

	must.NoError(err)
	is.Equal([]string{"story-aaa"}, ids)
}

func TestWalk_ExampleFromDoc(t *testing.T) {
	t.Parallel()

	// Example usage showing the typical pattern with mixed forms.
	fsys := memfs.New()
	testutil.NewFsBuilder(fsys).
		Dir("/project/.slope").
		File("/project/.slope/todo-gamma.md", "# Gamma task\n").
		Dir("/project/.slope/feature-auth").
		File("/project/.slope/feature-auth/README.md", "---\n_id: feature-auth\n---\n# Authentication\n").
		Dir("/project/.slope/story-init").
		File("/project/.slope/story-init/README.md", "---\n_id: story-init\n_archetype: story\n---\n# Initialize project\n")

	must := require.New(t)
	is := assert.New(t)

	var results []doc.Document
	err := doc.Walk(fsys, "/project/.slope", func(fsys billy.Filesystem, d doc.Document, err error) error {
		if err != nil {
			return fmt.Errorf("walk error: %w", err)
		}
		results = append(results, d)
		return nil
	})

	must.NoError(err)
	is.Len(results, 3)
	is.Equal("feature-auth", results[0].ID())
	is.Equal("feature", results[0].Archetype())
	is.Equal("story-init", results[1].ID())
	is.Equal("story", results[1].Archetype())
	is.Equal("todo-gamma", results[2].ID())
	is.Equal("todo", results[2].Archetype())
}
