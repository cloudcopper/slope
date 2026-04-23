package testutil

import (
	"io"
	"os"
	"testing"

	"github.com/go-git/go-billy/v5"
	"github.com/stretchr/testify/require"
)

func CheckFilesystem(t *testing.T, fsys billy.Filesystem) *FilesystemChecker {
	return &FilesystemChecker{t: t, fsys: fsys}
}

type FilesystemChecker struct {
	t    *testing.T
	fsys billy.Filesystem
}

// FileMustExist asserts that a file exists and is not a directory.
func (fc *FilesystemChecker) FileMustExist(path string) *FilesystemChecker {
	fc.t.Helper()
	info, err := fc.fsys.Stat(path)
	require.NoError(fc.t, err, "file %s should exist", path)
	require.False(fc.t, info.IsDir(), "path %s should be a file, not a directory", path)
	return fc
}

// FileMustNotExist asserts that a file does not exist.
func (fc *FilesystemChecker) FileMustNotExist(path string) *FilesystemChecker {
	fc.t.Helper()
	_, err := fc.fsys.Stat(path)
	require.True(fc.t, os.IsNotExist(err), "file %s should not exist", path)
	return fc
}

// DirMustExist asserts that a directory exists.
func (fc *FilesystemChecker) DirMustExist(path string) *FilesystemChecker {
	fc.t.Helper()
	info, err := fc.fsys.Stat(path)
	require.NoError(fc.t, err, "dir %s should exist", path)
	require.True(fc.t, info.IsDir(), "path %s should be a directory", path)
	return fc
}

// DirMustNotExist asserts that a directory does not exist.
func (fc *FilesystemChecker) DirMustNotExist(path string) *FilesystemChecker {
	fc.t.Helper()
	_, err := fc.fsys.Stat(path)
	require.True(fc.t, os.IsNotExist(err), "dir %s should not exist", path)
	return fc
}

// FileEqual asserts that a file contains exactly the expected content.
func (fc *FilesystemChecker) FileEqual(path, expectedContent string) *FilesystemChecker {
	fc.t.Helper()
	f, err := fc.fsys.Open(path)
	require.NoError(fc.t, err, "open file %s", path)
	defer f.Close()
	data, err := io.ReadAll(f)
	require.NoError(fc.t, err, "read file %s", path)
	require.Equal(fc.t, expectedContent, string(data), "file %s content mismatch", path)
	return fc
}

// FileNotEqual asserts that a file does NOT contain the expected content.
func (fc *FilesystemChecker) FileNotEqual(path, unexpectedContent string) *FilesystemChecker {
	fc.t.Helper()
	f, err := fc.fsys.Open(path)
	require.NoError(fc.t, err, "open file %s", path)
	defer f.Close()
	data, err := io.ReadAll(f)
	require.NoError(fc.t, err, "read file %s", path)
	require.NotEqual(fc.t, unexpectedContent, string(data), "file %s should not have this content", path)
	return fc
}

func (fc *FilesystemChecker) FilePermMustEqual(path string, expected os.FileMode) *FilesystemChecker {
	fc.t.Helper()
	info, err := fc.fsys.Stat(path)
	require.NoError(fc.t, err, "stat file %s", path)
	require.Equal(fc.t, expected, info.Mode().Perm(), "file %s permissions mismatch", path)
	return fc
}
func (fc *FilesystemChecker) FileMustSymlink(path string, expectedTarget string) *FilesystemChecker {
	fc.t.Helper()

	info, err := fc.fsys.Lstat(path)
	require.NoError(fc.t, err)
	require.True(fc.t, info.Mode()&os.ModeSymlink != 0)

	// Verify the link works
	target, err := fc.fsys.Readlink(path)
	require.NoError(fc.t, err)
	require.Equal(fc.t, expectedTarget, target)

	return fc
}
