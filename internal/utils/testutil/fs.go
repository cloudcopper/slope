package testutil

import (
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/util"
)

// NewFsBuilder creates a new FsBuilder with default 0o644 permissions.
func NewFsBuilder(fs billy.Filesystem) *FsBuilder {
	return &FsBuilder{
		fs:          fs,
		defaultPerm: 0o644,
	}
}

// FsBuilder provides a fluent API for building in-memory file trees.
type FsBuilder struct {
	fs          billy.Filesystem
	defaultPerm os.FileMode
}

// FileOption configures a file creation.
type FileOption func(*fileConfig)

type fileConfig struct {
	perm      os.FileMode
	symlinkTo string // if non-empty, create symlink instead of regular file
}

// WithPerm sets the file permission (default: 0o644).
func WithPerm(perm os.FileMode) FileOption {
	return func(c *fileConfig) { c.perm = perm }
}

// AsSymlink creates a symbolic link pointing to target.
func AsSymlink(target string) FileOption {
	return func(c *fileConfig) { c.symlinkTo = target }
}

// WithDefaultPerm sets the default permission for files.
func (b *FsBuilder) WithDefaultPerm(perm os.FileMode) *FsBuilder {
	b.defaultPerm = perm
	return b
}

// File creates a regular file with content. Parent directories are created implicitly.
// PANICS if file already exists or on any error.
func (b *FsBuilder) File(path, content string, opts ...FileOption) *FsBuilder {
	cfg := &fileConfig{perm: b.defaultPerm}
	for _, opt := range opts {
		opt(cfg)
	}

	// Check for overwrite - panic if exists
	if _, err := b.fs.Stat(path); err == nil {
		panicf("can not create file %s: already exists", path)
	}

	if cfg.symlinkTo != "" {
		// Create symlink
		if err := b.fs.Symlink(cfg.symlinkTo, path); err != nil {
			panicf("can not create file %s: %v", path, err)
		}
	} else {
		// Create regular file
		if err := util.WriteFile(b.fs, path, []byte(content), cfg.perm); err != nil {
			panicf("can not create file %s: %v", path, err)
		}
	}
	return b
}

// UpdateFile updates an existing file with new content.
// PANICS if file does not exist or on any error.
func (b *FsBuilder) UpdateFile(path, content string, opts ...FileOption) *FsBuilder {
	cfg := &fileConfig{perm: b.defaultPerm}
	for _, opt := range opts {
		opt(cfg)
	}

	// Check for existence - panic if does not exist
	if _, err := b.fs.Stat(path); err != nil {
		panicf("can not update file %s: does not exist", path)
	}

	// Create regular file
	if err := util.WriteFile(b.fs, path, []byte(content), cfg.perm); err != nil {
		panicf("can not update file %s: %v", path, err)
	}

	return b
}

// Dir creates an empty directory. PANICS if directory already exists or on any error.
func (b *FsBuilder) Dir(path string) *FsBuilder {
	// Check for overwrite - panic if exists
	if _, err := b.fs.Stat(path); err == nil {
		panicf("can not create dir %s: already exists", path)
	}
	if err := b.fs.MkdirAll(path, 0o755); err != nil {
		panicf("can not create dir %s: %v", path, err)
	}
	return b
}

// Symlink creates a symbolic link (alternative to File with AsSymlink).
func (b *FsBuilder) Symlink(path, target string) *FsBuilder {
	return b.File(path, "", AsSymlink(target))
}
