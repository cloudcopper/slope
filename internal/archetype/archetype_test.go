package archetype_test

import (
	"errors"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/egorse/slope/internal/archetype"
	"github.com/egorse/slope/internal/doc"
	"github.com/egorse/slope/internal/utils/testutil"
)

func TestFind_ProjectLocalFound(t *testing.T) {
	t.Parallel()

	const src = `---
title: Story baseline
---
# Story Template

Default content.
`
	fs := memfs.New()
	testutil.NewFsBuilder(fs).
		Dir(".slope/archetype").
		File(".slope/archetype/story.md", src)

	must := require.New(t)
	is := assert.New(t)

	d, err := archetype.Find(fs, ".slope", "/home/user", "story")
	must.NoError(err)
	is.Equal(".slope/archetype/story.md", d.Filepath)
	is.Equal("Story baseline", d.Metadata["title"])
	is.Equal("story", d.Archetype())
}

func TestFind_ProjectLocalTOML(t *testing.T) {
	t.Parallel()

	const src = `+++
title = "Feature baseline"
+++
# Feature Template
`
	fs := memfs.New()
	testutil.NewFsBuilder(fs).
		Dir(".slope/archetype").
		File(".slope/archetype/feature.md", src)

	must := require.New(t)
	is := assert.New(t)

	d, err := archetype.Find(fs, ".slope", "/home/user", "feature")
	must.NoError(err)
	is.Equal("Feature baseline", d.Metadata["title"])
	is.Equal("feature", d.Archetype())
}

func TestFind_ProjectLocalNotFound(t *testing.T) {
	t.Parallel()

	// Project-local not found — falls through to user-global.
	// User-global also not found — returns ErrOpen wrapping os.ErrNotExist.
	fs := memfs.New()
	testutil.NewFsBuilder(fs).
		Dir(".slope")

	must := require.New(t)
	is := assert.New(t)

	_, err := archetype.Find(fs, ".slope", "/home/user", "missing")
	must.Error(err)
	var target *doc.ErrOpen
	must.True(errors.As(err, &target))
	is.Equal("/home/user/.config/slope/archetype/missing.md", target.Filepath)
}

func TestFind_ProjectLocalMalformedFrontmatter(t *testing.T) {
	t.Parallel()

	const src = `---
meta:
  nested: value
---
body
`
	fs := memfs.New()
	testutil.NewFsBuilder(fs).
		Dir(".slope/archetype").
		File(".slope/archetype/story.md", src)

	must := require.New(t)
	is := assert.New(t)

	_, err := archetype.Find(fs, ".slope", "/home/user", "story")
	must.Error(err)
	var target *doc.ErrComplexMetadata
	must.True(errors.As(err, &target))
	is.Equal(".slope/archetype/story.md", target.Filepath)
	is.Equal("meta", target.Key)
}

func TestFind_ProjectLocalIsDirectory(t *testing.T) {
	t.Parallel()

	// Project-local path is a directory, not a file — ErrOpen is returned
	// (we do NOT fall through to user-global for non-NotExist errors).
	fs := memfs.New()
	testutil.NewFsBuilder(fs).
		Dir(".slope/archetype").
		Dir(".slope/archetype/story.md") // directory, not file

	must := require.New(t)
	is := assert.New(t)

	_, err := archetype.Find(fs, ".slope", "/home/user", "story")
	must.Error(err)
	var target *doc.ErrOpen
	must.True(errors.As(err, &target))
	is.Equal(".slope/archetype/story.md", target.Filepath)
}

func TestFind_UserGlobalFound(t *testing.T) {
	t.Parallel()

	const src = `---
title: User global archetype
---
# Global Template
`
	fs := memfs.New()
	testutil.NewFsBuilder(fs).
		Dir(".slope").
		Dir("/home/user/.config/slope/archetype").
		File("/home/user/.config/slope/archetype/todo.md", src)

	must := require.New(t)
	is := assert.New(t)

	d, err := archetype.Find(fs, ".slope", "/home/user", "todo")
	must.NoError(err)
	is.Equal("/home/user/.config/slope/archetype/todo.md", d.Filepath)
	is.Equal("User global archetype", d.Metadata["title"])
}

func TestFind_UserGlobalNotFound(t *testing.T) {
	t.Parallel()

	// Both project-local and user-global are missing.
	fs := memfs.New()
	testutil.NewFsBuilder(fs).
		Dir(".slope")

	must := require.New(t)
	is := assert.New(t)

	_, err := archetype.Find(fs, ".slope", "/home/user", "missing")
	must.Error(err)
	var target *doc.ErrOpen
	must.True(errors.As(err, &target))
	is.Equal("/home/user/.config/slope/archetype/missing.md", target.Filepath)
}

func TestFind_UserGlobalMalformedFrontmatter(t *testing.T) {
	t.Parallel()

	const src = `---
meta:
  nested: value
---
body
`
	fs := memfs.New()
	testutil.NewFsBuilder(fs).
		Dir(".slope").
		Dir("/home/user/.config/slope/archetype").
		File("/home/user/.config/slope/archetype/story.md", src)

	must := require.New(t)
	is := assert.New(t)

	_, err := archetype.Find(fs, ".slope", "/home/user", "story")
	must.Error(err)
	var target *doc.ErrComplexMetadata
	must.True(errors.As(err, &target))
	is.Equal("/home/user/.config/slope/archetype/story.md", target.Filepath)
}

func TestFind_EmptyName(t *testing.T) {
	t.Parallel()

	fs := memfs.New()
	testutil.NewFsBuilder(fs).
		Dir(".slope")

	must := require.New(t)
	is := assert.New(t)

	_, err := archetype.Find(fs, ".slope", "/home/user", "")
	must.Error(err)
	var target *doc.ErrOpen
	must.True(errors.As(err, &target))
	is.Equal("/home/user/.config/slope/archetype/.md", target.Filepath)
}

func TestFind_NestedProjectLocal(t *testing.T) {
	t.Parallel()

	const src = `---
title: Deep archetype
---
# Deep
`
	fs := memfs.New()
	testutil.NewFsBuilder(fs).
		Dir(".slope/archetype").
		File(".slope/archetype/story.md", src)

	must := require.New(t)
	is := assert.New(t)

	d, err := archetype.Find(fs, ".slope", "/home/user", "story")
	must.NoError(err)
	is.Equal("Deep archetype", d.Metadata["title"])
}

func TestFind_ProjectLocalPrecedence(t *testing.T) {
	t.Parallel()

	// Both project-local and user-global exist — project-local wins.
	localSrc := `---
title: Local
---
# Local
`
	globalSrc := `---
title: Global
---
# Global
`
	fs := memfs.New()
	testutil.NewFsBuilder(fs).
		Dir(".slope/archetype").
		File(".slope/archetype/story.md", localSrc).
		Dir("/home/user/.config/slope/archetype").
		File("/home/user/.config/slope/archetype/story.md", globalSrc)

	must := require.New(t)
	is := assert.New(t)

	d, err := archetype.Find(fs, ".slope", "/home/user", "story")
	must.NoError(err)
	is.Equal("Local", d.Metadata["title"])
}

func TestFind_UserGlobalPrecedence(t *testing.T) {
	t.Parallel()

	// Project-local not found — user-global is used.
	const globalSrc = `---
title: Global
---
# Global
`
	fs := memfs.New()
	testutil.NewFsBuilder(fs).
		Dir(".slope").
		Dir("/home/user/.config/slope/archetype").
		File("/home/user/.config/slope/archetype/story.md", globalSrc)

	must := require.New(t)
	is := assert.New(t)

	d, err := archetype.Find(fs, ".slope", "/home/user", "story")
	must.NoError(err)
	is.Equal("Global", d.Metadata["title"])
}

func TestFind_UserGlobalReadError(t *testing.T) {
	t.Parallel()

	// Project-local doesn't exist, user-global path is a directory.
	// This tests that a read error on user-global is returned as-is.
	fs := memfs.New()
	testutil.NewFsBuilder(fs).
		Dir(".slope").
		Dir("/home/user/.config/slope/archetype/story.md") // directory, not file

	must := require.New(t)
	is := assert.New(t)

	_, err := archetype.Find(fs, ".slope", "/home/user", "story")
	must.Error(err)
	// The error is from doc.NewFromFile trying to open a directory.
	var target *doc.ErrOpen
	must.True(errors.As(err, &target))
	is.Equal("/home/user/.config/slope/archetype/story.md", target.Filepath)
}

func TestFind_RootWithTrailingSlash(t *testing.T) {
	t.Parallel()

	const src = `---
title: Trailing slash test
---
# Trailing
`
	fs := memfs.New()
	testutil.NewFsBuilder(fs).
		Dir(".slope/archetype").
		File(".slope/archetype/story.md", src)

	must := require.New(t)
	is := assert.New(t)

	d, err := archetype.Find(fs, ".slope/", "/home/user", "story")
	must.NoError(err)
	is.Equal("Trailing slash test", d.Metadata["title"])
}

func TestFind_HomeWithTrailingSlash(t *testing.T) {
	t.Parallel()

	const src = `---
title: Home slash test
---
# Home
`
	fs := memfs.New()
	testutil.NewFsBuilder(fs).
		Dir(".slope").
		Dir("/home/user/.config/slope/archetype").
		File("/home/user/.config/slope/archetype/story.md", src)

	must := require.New(t)
	is := assert.New(t)

	d, err := archetype.Find(fs, ".slope", "/home/user/", "story")
	must.NoError(err)
	is.Equal("Home slash test", d.Metadata["title"])
}

func TestFind_UserGlobalEmptyHome(t *testing.T) {
	t.Parallel()

	fs := memfs.New()
	testutil.NewFsBuilder(fs).
		Dir(".slope")

	must := require.New(t)
	is := assert.New(t)

	_, err := archetype.Find(fs, ".slope", "", "missing")
	must.Error(err)
	var target *doc.ErrOpen
	must.True(errors.As(err, &target))
	is.Equal(".config/slope/archetype/missing.md", target.Filepath)
}
