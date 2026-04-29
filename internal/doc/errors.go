package doc

import "fmt"

// ErrOpen is returned when a document file cannot be opened.
type ErrOpen struct {
	Filepath string
	err      error
}

func (e *ErrOpen) Error() string {
	return fmt.Sprintf("doc: open %q: %s", e.Filepath, e.err)
}

func (e *ErrOpen) Unwrap() error { return e.err }

// ErrRead is returned when a document file cannot be read after opening.
type ErrRead struct {
	Filepath string
	err      error
}

func (e *ErrRead) Error() string {
	return fmt.Sprintf("doc: read %q: %s", e.Filepath, e.err)
}

func (e *ErrRead) Unwrap() error { return e.err }

// ErrFrontmatter is returned when the frontmatter block cannot be decoded.
type ErrFrontmatter struct {
	Filepath string
	err      error
}

func (e *ErrFrontmatter) Error() string {
	return fmt.Sprintf("doc: frontmatter in %q: %s", e.Filepath, e.err)
}

func (e *ErrFrontmatter) Unwrap() error { return e.err }

// ErrComplexMetadata is returned when a frontmatter key holds a nested value
// (map or slice) instead of a flat scalar.
type ErrComplexMetadata struct {
	Filepath string
	Key      string
}

func (e *ErrComplexMetadata) Error() string {
	return fmt.Sprintf("doc: frontmatter key %q in %q has a complex value; only flat key-value pairs are supported", e.Key, e.Filepath)
}

// ErrRender is returned when the Markdown AST cannot be rendered.
type ErrRender struct {
	Filepath string
	err      error
}

func (e *ErrRender) Error() string {
	return fmt.Sprintf("doc: render %q: %s", e.Filepath, e.err)
}

func (e *ErrRender) Unwrap() error { return e.err }
