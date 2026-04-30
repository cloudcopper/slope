// Package archetype provides functions for resolving and loading archetype
// Markdown documents.  Each document may carry optional TOML or YAML
// frontmatter and a Markdown body.
package archetype

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5"

	"github.com/egorse/slope/internal/doc"
)

// archetypeDir is the subdirectory name under the .slope root where
// project-local archetype files are stored.
const archetypeDir = "archetype"

// Find resolves and loads an archetype by name.
//
// Resolution order (per IDEA.md §3.3):
//
//	1. <root>/archetype/<name>.md  (project-local)
//	2. <home>/.config/slope/archetype/<name>.md  (user-global)
//
// Both paths are resolved on the provided fsys.  The caller is responsible
// for providing a fsys that can serve both locations (e.g. a real osfs for
// production, or a combined memfs for tests).
//
// The returned Document is loaded via doc.NewFromFile, which handles
// frontmatter parsing and Markdown AST construction.
func Find(fsys billy.Filesystem, root, home, name string) (doc.Document, error) {
	// 1. Try project-local: <root>/archetype/<name>.md
	localPath := filepath.Join(root, archetypeDir, name+".md")
	d, err := doc.NewFromFile(fsys, localPath)
	if err == nil {
		return d, nil
	}

	// If the file doesn't exist, continue to user-global.
	// Any other error (read failure, malformed frontmatter) is returned as-is.
	if !errors.Is(err, os.ErrNotExist) {
		return doc.Document{}, err
	}

	// 2. Try user-global: <home>/.config/slope/archetype/<name>.md
	globalPath := filepath.Join(home, ".config", "slope", archetypeDir, name+".md")
	return doc.NewFromFile(fsys, globalPath)
}
