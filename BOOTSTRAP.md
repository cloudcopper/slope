# Bootstrap: cmd/bootstrap/main.go

Generate a single-file Go program at `cmd/bootstrap/main.go` — a throwaway MVP of the Slope CLI (see IDEA.md for full design context). This bootstrap is used for dogfooding while the real implementation is built separately in `cmd/slope/`.

## General

- Single-file Go program at `cmd/bootstrap/main.go`
- Run via `go run cmd/bootstrap/main.go <command> [flags] [args]`
- Stdlib only (no external dependencies)
- Finds `.slope/` directory by walking up from the current working directory; error if not found
- **Constraint**: Avoid regex usage, especially for frontmatter processing; use simple string operations instead

## Commands

### `add [--parent=<id>] <archetype> "<text>"`

Create a new ticket file.

- `<text>` is the full markdown body of the ticket (passed as a CLI argument string)
- **Slug generation**: extract from the first `#` heading found in `<text>`; if no heading exists, use the first sentence. Lowercase, spaces to hyphens, strip characters that are not alphanumeric or hyphens.
- **File path**: `.slope/<archetype>-<slug>.md` at the root, or inside the parent's directory if `--parent` is specified
- **Leaf-to-directory promotion**: if the resolved parent is currently a leaf file (`<archetype>-<slug>.md`), promote it to directory form (`<archetype>-<slug>/README.md`) before placing the new child inside. On promotion, `_id` is added to frontmatter (unless already present).
- **Collision**: if a file with the same name already exists in the target directory, print an error and exit non-zero — no auto-resolution
- Write `<text>` as-is into the created `.md` file

### `prompt <id>`

Generate a merged prompt for the given ticket and print it to stdout.

1. **Find ticket**: depth-first search of `.slope/` matching `<id>` against filename stems (e.g., `story-aaa` matches `story-aaa.md` or `story-aaa/README.md`) or `_id` frontmatter in `README.md` files. Error if not found or if multiple matches exist.
2. **Build ancestor chain**: walk the filesystem path from `.slope/` root down to the found ticket. Each directory level that contains a `README.md` file is an ancestor.
3. **Build file sequence**: for each level top-down, append the ticket file (for directory tickets, use `README.md`). For a 3-level chain `story-aaa/feature-bbb/todo-ccc`, the sequence is:
   - `story-aaa/README.md`, `feature-bbb/README.md`, `todo-ccc.md`
4. **Merge**: for each file in sequence, strip its frontmatter and append the remaining body to the output. The result is a simple concatenation of all ticket bodies from root to leaf, with no heading-based merge.
5. **Output**: print the concatenated markdown to stdout.

## Constraints

- **No regex usage**: Avoid complex pattern matching; use simple string operations where possible, especially when processing frontmatter or ticket IDs.

## Out of Scope

Do NOT implement any of the following — they are deferred to `cmd/slope/`:

- **Archetype support** (archetype resolution, archetype files in merge sequence)
- Frontmatter / metadata accumulation (except `_id` for promotion)
- Template engine pass or chapter renumbering
- Heading-based AST merge
- `list`, `verify`, `init`, or any other commands
- `_id` override or ID collision auto-resolution
- `_remove_chapters` or `_remove_keys` metadata directives
- `SLOPE_SCOPE_ID` environment variable
- Reading ticket text from stdin or `$EDITOR` — CLI argument only

# Final notes
- DO NOT USE git
- DO NOT USE /tmp - use ./tmp when need temporary files
- DO NOT ASK question - implement cmd/bootstrap/main.go
- Ensure the frontmatters are removed by 'prompt' command
 
