package fs

import (
	"errors"
	"path/filepath"

	"github.com/go-git/go-billy/v5"
)

// ErrRootNotFound is returned when the search target is not found walking up to /.
var ErrRootNotFound = errors.New("root not found")

// FindRoot walks up from cwd toward "/" looking for a file or directory named
// search. It returns the first directory that contains search, or ErrRootNotFound
// if the filesystem root is reached without a match.
func FindRoot(fsys billy.Filesystem, cwd string, search string) (string, error) {
	dir := filepath.Clean(cwd)
	for {
		_, err := fsys.Stat(filepath.Join(dir, search))
		if err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root.
			return "", ErrRootNotFound
		}
		dir = parent
	}
}
