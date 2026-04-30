package doc_test

import (
	"errors"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/egorse/slope/internal/doc"
	"github.com/egorse/slope/internal/utils/testutil"
)

func TestNewFromFile_NoFrontmatter(t *testing.T) {
	t.Parallel()

	const src = `# Hello

Body text.
`

	fs := memfs.New()
	testutil.NewFsBuilder(fs).File("ticket.md", src)

	must := require.New(t)
	is := assert.New(t)

	d, err := doc.NewFromFile(fs, "ticket.md")
	must.NoError(err)
	is.Equal("ticket.md", d.Filepath)
	is.Equal(doc.Metadata{}, d.Metadata)
	is.NotNil(d.AST)

	out, err := doc.RenderMarkdown(d)
	must.NoError(err)
	is.Equal(src, out)
}

func TestNewFromFile_YAMLFrontmatter(t *testing.T) {
	t.Parallel()

	const src = `---
_id: feature-42
title: My Feature
---
# Body
`
	fs := memfs.New()
	testutil.NewFsBuilder(fs).File("feature-42.md", src)

	must := require.New(t)
	is := assert.New(t)

	d, err := doc.NewFromFile(fs, "feature-42.md")
	must.NoError(err)
	is.Equal("feature-42", d.Metadata["_id"])
	is.Equal("My Feature", d.Metadata["title"])
	// AST must contain only the body — frontmatter node is removed by goldmark-frontmatter.
	out, err := doc.RenderMarkdown(d)
	must.NoError(err)
	is.Equal("# Body\n", out)
}

func TestNewFromFile_TOMLFrontmatter(t *testing.T) {
	t.Parallel()

	const src = `+++
_id = "story-7"
author = "alice"
+++
Hello TOML.
`
	fs := memfs.New()
	testutil.NewFsBuilder(fs).File("story-7.md", src)

	must := require.New(t)
	is := assert.New(t)

	d, err := doc.NewFromFile(fs, "story-7.md")
	must.NoError(err)
	is.Equal("story-7", d.Metadata["_id"])
	is.Equal("alice", d.Metadata["author"])
	out, err := doc.RenderMarkdown(d)
	must.NoError(err)
	is.Equal("Hello TOML.\n", out)
}

func TestNewFromFile_TOMLIntegerValue(t *testing.T) {
	t.Parallel()

	// TOML integers decode as int64.
	const src = `+++
count = 42
ratio = 1.5
+++
body
`
	fs := memfs.New()
	testutil.NewFsBuilder(fs).File("t.md", src)

	must := require.New(t)
	is := assert.New(t)

	d, err := doc.NewFromFile(fs, "t.md")
	must.NoError(err)
	is.Equal("42", d.Metadata["count"])
	is.Equal("1.5", d.Metadata["ratio"])
}

func TestNewFromFile_FrontmatterBoolAndNumber(t *testing.T) {
	t.Parallel()

	const src = `---
active: true
count: 3
ratio: 1.5
---
`
	fs := memfs.New()
	testutil.NewFsBuilder(fs).File("t.md", src)

	must := require.New(t)
	is := assert.New(t)

	d, err := doc.NewFromFile(fs, "t.md")
	must.NoError(err)
	is.Equal("true", d.Metadata["active"])
	is.Equal("3", d.Metadata["count"])
	is.Equal("1.5", d.Metadata["ratio"])
}

func TestNewFromFile_FrontmatterNilValue(t *testing.T) {
	t.Parallel()

	// YAML null → nil → should be stored as empty string.
	const src = `---
key: null
---
body
`
	fs := memfs.New()
	testutil.NewFsBuilder(fs).File("t.md", src)

	must := require.New(t)
	is := assert.New(t)

	d, err := doc.NewFromFile(fs, "t.md")
	must.NoError(err)
	is.Equal("", d.Metadata["key"])
}

func TestNewFromFile_NestedYAMLReturnsError(t *testing.T) {
	t.Parallel()

	const src = `---
meta:
  nested: value
---
body
`
	fs := memfs.New()
	testutil.NewFsBuilder(fs).File("t.md", src)

	must := require.New(t)
	is := assert.New(t)

	_, err := doc.NewFromFile(fs, "t.md")
	must.Error(err)
	var target *doc.ErrComplexMetadata
	must.True(errors.As(err, &target))
	is.Equal("t.md", target.Filepath)
	is.Equal("meta", target.Key)
}

func TestNewFromFile_NestedTOMLReturnsError(t *testing.T) {
	t.Parallel()

	const src = `+++
[section]
key = "value"
+++
body
`
	fs := memfs.New()
	testutil.NewFsBuilder(fs).File("t.md", src)

	must := require.New(t)
	is := assert.New(t)

	_, err := doc.NewFromFile(fs, "t.md")
	must.Error(err)
	var target *doc.ErrComplexMetadata
	must.True(errors.As(err, &target))
	is.Equal("t.md", target.Filepath)
	is.Equal("section", target.Key)
}

func TestNewFromFile_FileNotFound(t *testing.T) {
	t.Parallel()

	fs := memfs.New()
	must := require.New(t)
	is := assert.New(t)

	_, err := doc.NewFromFile(fs, "missing.md")
	must.Error(err)
	var target *doc.ErrOpen
	must.True(errors.As(err, &target))
	is.Equal("missing.md", target.Filepath)
	is.NotNil(target.Unwrap())
}

func TestNewFromFile_EmptyFile(t *testing.T) {
	t.Parallel()

	fs := memfs.New()
	testutil.NewFsBuilder(fs).File("empty.md", "")

	must := require.New(t)
	is := assert.New(t)

	d, err := doc.NewFromFile(fs, "empty.md")
	must.NoError(err)
	is.Equal(doc.Metadata{}, d.Metadata)
	is.NotNil(d.AST)

	out, err := doc.RenderMarkdown(d)
	must.NoError(err)
	is.Equal("\n", out)
}

func TestNewFromFile_FrontmatterNoClosingDelimiter(t *testing.T) {
	t.Parallel()

	// No closing --- and invalid YAML content → goldmark-frontmatter tries to
	// parse and returns an error (malformed frontmatter).
	const src = `---
title: Orphaned
body text
`
	fs := memfs.New()
	testutil.NewFsBuilder(fs).File("t.md", src)

	must := require.New(t)
	is := assert.New(t)

	_, err := doc.NewFromFile(fs, "t.md")
	must.Error(err)
	var target *doc.ErrFrontmatter
	must.True(errors.As(err, &target))
	is.Equal("t.md", target.Filepath)
	is.NotNil(target.Unwrap())
}

func TestDocument_ID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filepath string
		metadata doc.Metadata
		expected string
	}{
		{
			name:     "from _id metadata",
			filepath: "feature-old.md",
			metadata: doc.Metadata{"_id": "feature-new"},
			expected: "feature-new",
		},
		{
			name:     "from filename stem",
			filepath: "feature-slug.md",
			metadata: doc.Metadata{},
			expected: "feature-slug",
		},
		{
			name:     "README.md — derived from containing directory name",
			filepath: "story-abc/README.md",
			metadata: doc.Metadata{},
			expected: "story-abc",
		},
		{
			name:     "README.md — _id metadata takes precedence over directory name",
			filepath: "story-abc/README.md",
			metadata: doc.Metadata{"_id": "story-override"},
			expected: "story-override",
		},
		{
			name:     "README.md at root — falls back to stem",
			filepath: "README.md",
			metadata: doc.Metadata{},
			expected: "README",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			d := doc.Document{Filepath: tt.filepath, Metadata: tt.metadata}
			is.Equal(tt.expected, d.ID())
		})
	}
}

func TestDocument_Archetype(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filepath string
		metadata doc.Metadata
		expected string
	}{
		{
			name:     "from _archetype metadata",
			filepath: "feature-foo.md",
			metadata: doc.Metadata{"_archetype": "custom"},
			expected: "custom",
		},
		{
			name:     "first hyphen segment of filename stem",
			filepath: "story-my-ticket.md",
			metadata: doc.Metadata{},
			expected: "story",
		},
		{
			name:     "no hyphen in stem — whole stem is archetype",
			filepath: "standalone.md",
			metadata: doc.Metadata{},
			expected: "standalone",
		},
		{
			name:     "README.md — derived from containing directory name",
			filepath: "feature-bar/README.md",
			metadata: doc.Metadata{},
			expected: "feature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			d := doc.Document{Filepath: tt.filepath, Metadata: tt.metadata}
			is.Equal(tt.expected, d.Archetype())
		})
	}
}

func TestRenderMarkdown_YAML(t *testing.T) {
	t.Parallel()

	const src = `---
_id: todo-5
_archetype: todo
note: hello
---
# Task

Do things.
`
	fs := memfs.New()
	testutil.NewFsBuilder(fs).File("todo-5.md", src)

	must := require.New(t)
	is := assert.New(t)

	d, err := doc.NewFromFile(fs, "todo-5.md")
	must.NoError(err)
	is.Equal("todo-5", d.ID())
	is.Equal("todo", d.Archetype())

	out, err := doc.RenderMarkdown(d)
	must.NoError(err)
	is.Equal("# Task\n\nDo things.\n", out)
}

func TestRenderMarkdown_TOML(t *testing.T) {
	t.Parallel()

	const src = `+++
_id = "feature-99"
_archetype = "feature"
+++
Content.
`
	fs := memfs.New()
	testutil.NewFsBuilder(fs).File("feature-99.md", src)

	must := require.New(t)
	is := assert.New(t)

	d, err := doc.NewFromFile(fs, "feature-99.md")
	must.NoError(err)
	is.Equal("feature-99", d.ID())
	is.Equal("feature", d.Archetype())

	out, err := doc.RenderMarkdown(d)
	must.NoError(err)
	is.Equal("Content.\n", out)
}
