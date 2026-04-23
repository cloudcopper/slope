package testutil

import (
	"os"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/stretchr/testify/require"
)

// TestBuilderFileBasic tests basic file creation.
func TestBuilderFileBasic(t *testing.T) {
	fs := memfs.New()
	NewFsBuilder(fs).
		File("readme.txt", "hello world")

	CheckFilesystem(t, fs).
		FileMustExist("readme.txt").
		FileEqual("readme.txt", "hello world")
}

// TestBuilderFileNested tests nested directory file creation.
func TestBuilderFileNested(t *testing.T) {
	fs := memfs.New()
	NewFsBuilder(fs).
		File("a/b/c/deep.txt", "nested content")

	CheckFilesystem(t, fs).
		DirMustExist("a").
		DirMustExist("a/b").
		DirMustExist("a/b/c").
		FileMustExist("a/b/c/deep.txt").
		FileEqual("a/b/c/deep.txt", "nested content")
}

// TestBuilderDir tests directory creation.
func TestBuilderDir(t *testing.T) {
	fs := memfs.New()
	NewFsBuilder(fs).
		Dir("emptydir")

	CheckFilesystem(t, fs).
		DirMustExist("emptydir")
}

// TestBuilderDirNested tests nested directory creation.
func TestBuilderDirNested(t *testing.T) {
	fs := memfs.New()
	NewFsBuilder(fs).
		Dir("a/b/c1").
		Dir("a/b/c2")

	CheckFilesystem(t, fs).
		DirMustExist("a").
		DirMustExist("a/b").
		DirMustExist("a/b/c1").
		DirMustExist("a/b/c2")
}

// TestBuilderChained tests fluent chaining.
func TestBuilderChained(t *testing.T) {
	fs := memfs.New()
	NewFsBuilder(fs).
		File("file1.txt", "content1").
		File("file2.txt", "content2").
		Dir("subdir").
		File("subdir/file3.txt", "content3")

	CheckFilesystem(t, fs).
		FileMustExist("file1.txt").
		FileMustExist("file2.txt").
		DirMustExist("subdir").
		FileMustExist("subdir/file3.txt")
}

// TestBuilderFileOverwritePanic tests that overwriting a file panics.
func TestBuilderFileOverwritePanic(t *testing.T) {
	must := require.New(t)

	fs := memfs.New()
	NewFsBuilder(fs).
		File("file.txt", "original")

	must.Panics(func() {
		NewFsBuilder(fs).
			File("file.txt", "overwrite")
	})
}

// TestBuilderDirOverwritePanic tests that overwriting a directory panics.
func TestBuilderDirOverwritePanic(t *testing.T) {
	must := require.New(t)

	fs := memfs.New()
	NewFsBuilder(fs).
		Dir("dir")

	must.Panics(func() {
		NewFsBuilder(fs).
			Dir("dir")
	})
}

// TestBuilderFileOverwriteByDir tests that creating a dir where file exists panics.
func TestBuilderFileOverwriteByDir(t *testing.T) {
	must := require.New(t)

	fs := memfs.New()
	NewFsBuilder(fs).
		File("path.txt", "content")

	must.Panics(func() {
		NewFsBuilder(fs).
			Dir("path.txt")
	})
}

// TestBuilderDirOverwriteByFile tests that creating a file where dir exists panics.
func TestBuilderDirOverwriteByFile(t *testing.T) {
	must := require.New(t)

	fs := memfs.New()
	NewFsBuilder(fs).
		Dir("path")

	must.Panics(func() {
		NewFsBuilder(fs).
			File("path", "content")
	})
}

// TestBuilderWithPerm tests custom permissions.
func TestBuilderWithPerm(t *testing.T) {
	fs := memfs.New()
	NewFsBuilder(fs).File("permed.txt", "content", WithPerm(0o600))

	CheckFilesystem(t, fs).
		FilePermMustEqual("permed.txt", os.FileMode(0o600)).
		FileEqual("permed.txt", "content")
}

// TestBuilderDefaultPerm tests default permissions.
func TestBuilderDefaultPerm(t *testing.T) {
	fs := memfs.New()
	NewFsBuilder(fs).File("default.txt", "content")

	CheckFilesystem(t, fs).
		FilePermMustEqual("default.txt", os.FileMode(0o644)).
		FileEqual("default.txt", "content")
}

// TestBuilderWithDefaultPerm tests WithDefaultPerm option.
func TestBuilderWithDefaultPerm(t *testing.T) {
	fs := memfs.New()
	NewFsBuilder(fs).
		WithDefaultPerm(0o600).
		File("custom.txt", "content")

	CheckFilesystem(t, fs).
		FilePermMustEqual("custom.txt", os.FileMode(0o600)).
		FileEqual("custom.txt", "content")
}

// TestBuilderSymlink tests symlink creation via Symlink method.
func TestBuilderSymlink(t *testing.T) {
	fs := memfs.New()
	NewFsBuilder(fs).
		File("target.txt", "target content").
		Symlink("link.txt", "target.txt")

	CheckFilesystem(t, fs).
		FileMustExist("target.txt").
		FileEqual("target.txt", "target content").
		FileMustSymlink("link.txt", "target.txt").
		FileEqual("link.txt", "target content")
}

// TestBuilderSymlinkViaOption tests symlink creation via AsSymlink option.
func TestBuilderSymlinkViaOption(t *testing.T) {
	fs := memfs.New()
	NewFsBuilder(fs).
		File("target.txt", "target content").
		File("link.txt", "", AsSymlink("target.txt"))

	CheckFilesystem(t, fs).
		FileMustExist("target.txt").
		FileEqual("target.txt", "target content").
		FileMustSymlink("link.txt", "target.txt").
		FileEqual("link.txt", "target content")
}

// TestBuilderFileEmptyContent tests creating file with empty content.
func TestBuilderFileEmptyContent(t *testing.T) {
	fs := memfs.New()
	NewFsBuilder(fs).
		File("empty.txt", "")

	CheckFilesystem(t, fs).
		FileMustExist("empty.txt").
		FileEqual("empty.txt", "")
}

// TestAssertFileMustNotExist tests FileMustNotExist assertion.
func TestAssertFileMustNotExist(t *testing.T) {
	fs := memfs.New()
	NewFsBuilder(fs).File("exists.txt", "content")
	CheckFilesystem(t, fs).FileMustNotExist("nonexistent.txt")
}

// TestAssertDirMustNotExist tests DirMustNotExist assertion.
func TestAssertDirMustNotExist(t *testing.T) {
	fs := memfs.New()
	NewFsBuilder(fs).Dir("exists")
	CheckFilesystem(t, fs).DirMustNotExist("nonexistent")
}

// TestAssertFileNotEqual tests FileNotEqual assertion.
func TestAssertFileNotEqual(t *testing.T) {
	fs := memfs.New()
	NewFsBuilder(fs).File("file.txt", "actual content")
	CheckFilesystem(t, fs).FileNotEqual("file.txt", "different content")
}
