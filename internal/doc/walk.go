package doc

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5"
)

// WalkFunc is the callback invoked by Walk for each loaded document.
//
// The fsys parameter is the filesystem passed to Walk.  The doc parameter
// contains the loaded document when err is nil.  When err is non-nil, doc
// holds the zero value and err describes the loading failure.
//
// Returning a non-nil error from WalkFunc stops the walk and is returned
// by Walk itself.
type WalkFunc func(fsys billy.Filesystem, doc Document, err error) error

// ErrWalkDir is returned when the root directory cannot be opened.
type ErrWalkDir struct {
	Path string
}

func (e *ErrWalkDir) Error() string {
	return fmt.Sprintf("doc: walk directory %q: directory not found", e.Path)
}

// Walk walks the directory tree rooted at root, invoking fn for each
// loadable document found at the first level.
//
// Walk scans only one level deep:
//
//   - Regular .md files (except README.md) are loaded directly.
//   - Directories are checked for a README.md child; if present, that
//     README.md is loaded as a document.
//
// Hidden entries (names starting with ".") are skipped.  Non-.md files
// and directories without README.md are silently ignored.
//
// If fn returns an error, Walk stops immediately and returns that error.
// If root does not exist, Walk returns an ErrWalkDir.
func Walk(fsys billy.Filesystem, root string, fn WalkFunc) error {
	root = filepath.Clean(root)

	entries, err := fsys.ReadDir(root)
	if err != nil {
		return &ErrWalkDir{Path: root}
	}

	for _, entry := range entries {
		name := entry.Name()

		// Skip hidden entries.
		if strings.HasPrefix(name, ".") {
			continue
		}

		path := filepath.Join(root, name)

		if entry.IsDir() {
			// Load README.md from the directory (promoted ticket).
			readmePath := filepath.Join(path, "README.md")
			if _, err := fsys.Stat(readmePath); err != nil {
				// No README.md in this directory; skip.
				continue
			}
			doc, err := NewFromFile(fsys, readmePath)
			if err := fn(fsys, doc, err); err != nil {
				return err
			}
			continue
		}
		if strings.HasSuffix(name, ".md") && name != "README.md" {
			// Load regular .md file (not README.md).
			doc, err := NewFromFile(fsys, path)
			if err := fn(fsys, doc, err); err != nil {
				return err
			}
		}
		// Non-.md files are silently skipped.
	}

	return nil
}
