package doc

import (
	"github.com/Kunde21/markdownfmt/v3/markdown"
	"github.com/yuin/goldmark"
	"go.abhg.dev/goldmark/frontmatter"
)

// MarkdownParser is the shared goldmark instance used for parsing Markdown
// documents with frontmatter support.
//
// We intentionally do NOT use frontmatter.SetMetadata mode.  That mode
// registers a MetaTransformer which decodes frontmatter into ast.Document.Meta()
// but silently discards any decode error (the transformer just returns on
// failure).  We need to surface decode errors — the spec requires returning
// an error for broken frontmatter.  Instead we retrieve the raw *frontmatter.Data
// from the parser.Context via frontmatter.Get and call Decode ourselves, where
// we can inspect and propagate the error.
//
// goldmark-frontmatter removes the frontmatter block from the AST during
// parsing, so the resulting ast.Node contains only the Markdown body.
var MarkdownParser = goldmark.New(
	goldmark.WithExtensions(
		&frontmatter.Extender{},
	),
)

// MarkdownRenderer is the shared markdownfmt renderer used to render a
// Markdown AST back to normalised Markdown text.
var MarkdownRenderer = markdown.NewRenderer()
