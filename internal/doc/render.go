package doc

import "bytes"

// RenderMarkdown renders d's Markdown body AST back to normalised Markdown
// text using the shared MarkdownRenderer.
//
// The returned string contains only the body — frontmatter is not re-emitted,
// because the AST produced by NewFromFile has the frontmatter node already
// removed by goldmark-frontmatter.
func RenderMarkdown(d Document) (string, error) {
	var buf bytes.Buffer
	if err := MarkdownRenderer.Render(&buf, d.Source, d.AST); err != nil {
		return "", &ErrRender{Filepath: d.Filepath, err: err}
	}
	return buf.String(), nil
}
