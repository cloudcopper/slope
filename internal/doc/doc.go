// Package doc provides types and constructors for working with slope Markdown
// documents (tickets and archetypes).  Each document may carry optional
// TOML or YAML frontmatter and a Markdown body.
package doc

import (
	"io"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// Metadata holds the flat key-value pairs parsed from a document's
// frontmatter.  Complex/nested values are rejected at parse time.
type Metadata map[string]string

// Document represents a single slope Markdown file (ticket or archetype).
type Document struct {
	// Fsys is the filesystem from which this document was loaded.
	Fsys billy.Filesystem

	// Filepath is the path as passed to NewFromFile.
	Filepath string

	// Metadata contains the flat key-value pairs from frontmatter (never nil).
	Metadata Metadata

	// Source is the raw file bytes.  AST segment offsets refer into this slice.
	Source []byte

	// AST is the parsed Markdown body with the frontmatter block already
	// removed by goldmark-frontmatter.  Use Source alongside the AST when
	// rendering or walking nodes.
	AST ast.Node
}

// ID returns the canonical ticket ID for this document.
//
// Resolution order (per IDEA.md §3.2):
//  1. The "_id" metadata key, if present and non-empty.
//  2. For README.md: the containing directory name.
//  3. The filename stem (base name without ".md" extension).
func (d Document) ID() string {
	if id, ok := d.Metadata["_id"]; ok && id != "" {
		return id
	}
	base := filepath.Base(d.Filepath)
	if base == "README.md" {
		dir := filepath.Base(filepath.Dir(d.Filepath))
		if dir != "." && dir != "" {
			return dir
		}
	}
	return strings.TrimSuffix(base, ".md")
}

// Archetype returns the ticket archetype for this document.
//
// Resolution order (per IDEA.md §3.2):
//  1. The "_archetype" metadata key, if present and non-empty.
//  2. The first hyphen-delimited segment of ID().
func (d Document) Archetype() string {
	if a, ok := d.Metadata["_archetype"]; ok && a != "" {
		return a
	}
	stem := d.ID()
	if idx := strings.IndexByte(stem, '-'); idx != -1 {
		return stem[:idx]
	}
	return stem
}

// NewFromFile loads a Document from filename on fsys.
// It auto-detects TOML (+++) or YAML (---) frontmatter.
// Returns an error if the file cannot be read, if frontmatter is malformed,
// or if it contains nested (non-flat) values.
func NewFromFile(fsys billy.Filesystem, filename string) (Document, error) {
	f, err := fsys.Open(filename)
	if err != nil {
		return Document{}, &ErrOpen{Filepath: filename, err: err}
	}
	defer f.Close() //nolint:errcheck

	src, err := io.ReadAll(f)
	if err != nil {
		return Document{}, &ErrRead{Filepath: filename, err: err}
	}

	ctx := parser.NewContext()
	node := MarkdownParser.Parser().Parse(text.NewReader(src), parser.WithContext(ctx))

	meta, err := parseFrontmatter(ctx, filename)
	if err != nil {
		return Document{}, err
	}

	return Document{
		Fsys:     fsys,
		Filepath: filename,
		Metadata: meta,
		Source:   src,
		AST:      node,
	}, nil
}
