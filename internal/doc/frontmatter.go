package doc

import (
	"fmt"

	"github.com/yuin/goldmark/parser"
	"go.abhg.dev/goldmark/frontmatter"
)

// parseFrontmatter extracts a flat Metadata map from the parser context.
// Returns an error if the frontmatter is malformed or contains nested values.
func parseFrontmatter(ctx parser.Context, filename string) (Metadata, error) {
	meta := Metadata{}
	data := frontmatter.Get(ctx)
	if data == nil {
		return meta, nil
	}

	raw := map[string]any{}
	if err := data.Decode(&raw); err != nil {
		return nil, &ErrFrontmatter{Filepath: filename, err: err}
	}
	for k, v := range raw {
		sv, ok := scalarString(v)
		if !ok {
			return nil, &ErrComplexMetadata{Filepath: filename, Key: k}
		}
		meta[k] = sv
	}
	return meta, nil
}

// scalarString converts a frontmatter value to a string.
// Returns (value, true) for scalar types; ("", false) for maps/slices.
func scalarString(v any) (string, bool) {
	switch t := v.(type) {
	case string:
		return t, true
	case bool:
		return fmt.Sprintf("%v", t), true
	case int:
		return fmt.Sprintf("%d", t), true
	case int64:
		return fmt.Sprintf("%d", t), true
	case float64:
		return fmt.Sprintf("%g", t), true
	case nil:
		return "", true
	default:
		// TODO Shall it panic instead?
		_ = t
		return "", false
	}
}
