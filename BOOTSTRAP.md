# Bootstrap: cmd/bootstrap/main.go

Generate a single-file Go program at `cmd/bootstrap/main.go` — a throwaway MVP of the Slope CLI (see IDEA.md for full design context). This bootstrap is used for dogfooding while the real implementation is built separately in `cmd/slope/`.

## General

- Single-file Go program at `cmd/bootstrap/main.go`
- Run via `go run cmd/bootstrap/main.go <command> [flags] [args]`
- Only external dependency: `github.com/yuin/goldmark` (plus stdlib)
- Finds `.slope/` directory by walking up from the current working directory; error if not found

## Commands

### `add [--parent=<id>] <archetype> "<text>"`

Create a new ticket file.

- `<text>` is the full markdown body of the ticket (passed as a CLI argument string)
- **Slug generation**: extract from the first `#` heading found in `<text>`; if no heading exists, use the first sentence. Lowercase, spaces to hyphens, strip characters that are not alphanumeric or hyphens.
- **File path**: `.slope/<archetype>-<slug>.md` at the root, or inside the parent's directory if `--parent` is specified
- **Leaf-to-directory promotion**: if the resolved parent is currently a leaf file (`<archetype>-<slug>.md`), promote it to directory form (`<archetype>-<slug>/<archetype>-<slug>.md`) before placing the new child inside
- **Collision**: if a file with the same name already exists in the target directory, print an error and exit non-zero — no auto-resolution
- Write `<text>` as-is into the created `.md` file

### `prompt <id>`

Generate a merged prompt for the given ticket and print it to stdout.

1. **Find ticket**: breadth-first search of `.slope/` matching `<id>` against filename stems (e.g., `story-aaa` matches `story-aaa.md` or `story-aaa/story-aaa.md`). Error if not found or if multiple matches exist.
2. **Build ancestor chain**: walk the filesystem path from `.slope/` root down to the found ticket. Each directory level that contains a self-named `.md` file is an ancestor.
3. **Resolve archetypes**: for each ticket in the chain, determine its archetype (the part before the first `-` in the filename stem). Look for the archetype file in order:
   - `.slope/archetype/<archetype>.md`
   - `~/.config/slope/archetype/<archetype>.md`
   - If neither exists, skip (no archetype contribution for that level)
4. **Build merge sequence**: for each level top-down: archetype file (if found), then the ticket file. For a 3-level chain `story-aaa/feature-bbb/todo-ccc`, the sequence is:
   - `archetype/story.md`, `story-aaa.md`, `archetype/feature.md`, `feature-bbb.md`, `archetype/todo.md`, `todo-ccc.md`
5. **Heading-based AST merge**: parse each file with goldmark into an AST. Merge sequentially into a single document:
   - Headings at all levels (`#` through `######`) define chapters. A chapter is a heading plus all content until the next heading of equal or higher level.
   - **New heading text**: insert the chapter into the merged AST
   - **Existing heading text** (exact match): replace the entire chapter body with the new version
   - **Existing heading with empty body**: remove that chapter from the merged AST
   - Content before the first heading merges the same way: later file's preamble replaces earlier
6. **Output**: render the merged AST back to Markdown and print to stdout

## Out of Scope

Do NOT implement any of the following — they are deferred to `cmd/slope/`:

- Frontmatter / metadata parsing or accumulation
- `list`, `verify`, `init`, or any other commands
- `_id` override or ID collision auto-resolution
- Template engine pass or chapter renumbering
- `_remove_chapters` or `_remove_keys` metadata directives
- `SLOPE_SCOPE_ID` environment variable
- Reading ticket text from stdin or `$EDITOR` — CLI argument only
