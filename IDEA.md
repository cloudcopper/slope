# Slope — Hierarchical Ticket System with Prompt Generation

## 1. Overview

Slope is a dual-purpose CLI tool designed for AI-assisted coding workflows. It provides:

1. A filesystem-based hierarchical ticket system stored within a repository
2. Templated prompt generation from selected tickets, using the hierarchy as a taxonomy for prompt composition

The key insight is that the ticket hierarchy itself defines the structure and content of generated prompts. Parent tickets and archetype templates contribute baseline chapters that child tickets inherit, override, or extend — eliminating the need to repeat instructions at every level.

## 2. Prior Art and References

- [backlogmd/backlogmd](https://github.com/backlogmd/backlogmd) — Markdown-based backlog management
- [MrLesk/Backlog.md](https://github.com/MrLesk/Backlog.md) — Backlog tracking in Markdown files
- [yuin/goldmark](https://github.com/yuin/goldmark) — Markdown parser for Go, used as reference for AST-based markdown processing

Slope differs from the above by introducing **prompt generation** as a first-class operation over the ticket hierarchy.

## 3. Concepts

### 3.1. Ticket

The fundamental unit. A ticket is a Markdown file with optional frontmatter metadata (TOML or YAML). Every ticket has:

- An **archetype** — a classification such as `story`, `feature`, `todo`, `issue` and etc.
- A **slug** — a human-readable identifier
- An **ID** — by default `<archetype>-<slug>`, guaranteed unique within the `.slope/` tree (see Section 3.4)
- Optional **frontmatter** — key-value metadata. Keys prefixed with `_` are reserved for slope-internal use (e.g., `_id`, `_allowed_children`, `_remove_keys`). All other keys are user-defined and available to templates.
- A **body** — Markdown content, optionally containing template directives

### 3.2. Hierarchy

The hierarchy is expressed through the filesystem. Tickets are stored under a `.slope/` directory at the repository root. The tool locates `.slope/` by traversing upward from the current working directory.

A ticket may exist as a file (leaf) or a directory (with children):

```
.slope/
  story-aaa.md                        # leaf ticket
  story-bbb/                          # ticket with children
    story-bbb.md                      # the ticket itself
    feature-b11.md                    # child (leaf)
    feature-b12.md                    # child (leaf)
    feature-b13/                      # child with own children
      feature-b13.md
      todo-c20.md
```

**Naming rules:**

- Default file ticket: `<archetype>-<slug>.md`
- Default directory ticket: `<archetype>-<slug>/<archetype>-<slug>.md` — the directory name and ticket filename must be the same. When `_id` overrides the canonical ID, both the directory and filename are renamed to match (e.g., `todo-101/todo-101.md`).
- Adding a child to a leaf ticket using ```slope add ...``` promotes it to directory form automatically
- The id may be overwritten by `_id` metadata
- The archetype may be overwritten by `_archetype` metadata

### 3.3. Archetype

An archetype is a Markdown file that provides baseline content for a given ticket type. Archetypes are resolved in order:

1. `.slope/archetype/<name>.md` (project-local)
2. `~/.config/slope/archetype/<name>.md` (user-global)

Archetypes are not templates in a special sense — they are plain Markdown files processed through the same merge mechanism as tickets. An archetype is merged into the AST immediately before its corresponding ticket, providing default chapters that the ticket may override/remove. If no archetype file exists for a given type, the merge chain includes only the ticket file for that level.

### 3.4. Ticket ID

Every ticket ID must be unique within the `.slope/` tree. The default ID is `<archetype>-<slug>`, derived from the filename.

**ID collision resolution at creation time:**

1. Generate default ID: `<archetype>-<slug>`
2. If unique across the tree — use as-is, no `_id` metadata needed
3. If ID collides — auto-generate `_id = <archetype>-<N>` where `<N>` is a globally incremented counter (scan all existing filenames and `_id` values for the highest numeric suffix, increment by one). The `_id` becomes the canonical ticket ID.
4. If the resulting filename also collides within the target directory — the filename becomes `<archetype>-<N>.md` matching the `_id`

The global counter is shared across all archetypes (e.g., `todo-101`, `feature-102`, `story-103`).

**Lookup:** `slope prompt <id>` matches against ticket IDs (either filename stem or `_id` if present) via breadth-first search. If a duplicate is detected (e.g., due to manual editing), a warning or error is raised. `slope verify` checks ID uniqueness across the tree.

## 4. Prompt Generation Pipeline

Given a ticket ID, the prompt generation pipeline produces a single Markdown document by merging content from the full ancestor chain and their archetypes.

### 4.1. Merge Chain

For a ticket at path `story-aaa/feature-bbb/todo-ccc`, the merge reads files in this order:

1. `archetype/story.md` (archetype baseline)
2. `story-aaa.md` (ticket content)
3. `archetype/feature.md` (archetype baseline)
4. `feature-bbb.md` (ticket content)
5. `archetype/todo.md` (archetype baseline)
6. `todo-ccc.md` (ticket content)

### 4.2. AST Merge Rules

Each file is parsed into a Markdown AST. Headings at all levels (`#`, `##`, `###`, etc.) define a tree of chapters. Merge proceeds by heading text:

- **New heading** — the chapter is inserted into the AST, ordered by its numeric prefix (if present) or appended after the last numbered chapter
- **Existing heading** — the chapter body is fully replaced by the later version. If the new chapter has an empty body, the chapter is removed from the AST. Chapters may also be removed via metadata: `_remove_chapters = ["Testing", "Acceptance Criteria"]`.

This allows parent tickets and archetypes to define default content that children selectively override/remove.

### 4.3. Metadata Accumulation

Frontmatter from each file in the merge chain is accumulated into a flat metadata map. On key collision, the later (child) value wins. The full per-ticket metadata chain is preserved for advanced use. A ticket may specify `_remove_keys = ["key1", "key2"]` to delete keys from the accumulated metadata at the point that ticket is processed.

### 4.4. Rendering

The pipeline executes in order:

1. **Merge** — walk the chain top-down, parse and merge each file into a combined AST
2. **Render** — serialize the merged AST back to Markdown
3. **Template** — pass the rendered Markdown through a template engine with the accumulated metadata as context
4. **Renumber** — reparse and sequentially renumber chapter headings

Template directives may appear in any source file (archetype or ticket). They are inert during the merge phase and resolved only at the template step. The exact template context shape (e.g., flat metadata map, per-level chain, current ticket fields) is to be defined during implementation.

## 5. Parent Resolution

When creating or operating on tickets, the parent context is resolved in order:

1. `--parent=<id>` CLI flag (explicit)
2. `SLOPE_SCOPE_ID` environment variable (set by external tooling or wrapping commands)
3. Root of `.slope/` (default)

The environment variable enables agent-aware workflows where a parent process sets the scope for child invocations.

## 6. CLI Interface (v0)

### 6.1. `slope add [--parent=<id>] <archetype> ["<text>"]`

Create a new ticket. Text may be provided as:

- A CLI argument (inline)
- Piped via stdin
- Entered in `$EDITOR` if neither argument nor stdin is provided

The ticket ID is derived from the archetype and a slug generated from the title. An explicit `--id` flag overrides the generated slug. ID uniqueness is guaranteed per Section 3.4.

### 6.2. `slope prompt [--debug] <id>`

Generate a prompt for the given ticket by executing the full merge and render pipeline (Section 4). Output is written to stdout.

The `--debug` flag produces a detailed processing log suitable for validating correctness against golden files.

### 6.3. `slope list [--parent=<id>]`

Display the ticket hierarchy as a tree. Optionally scoped to a subtree.

### 6.4. `slope verify`

Check consistency of the `.slope/` tree:

- ID uniqueness
- File/directory naming convention compliance
- Orphaned files
- Archetype resolution

## 7. Future Considerations

The following are explicitly deferred from v0:

- **Archetype constraints** — frontmatter rules such as `_allowed_children = ["feature", "task"]` to enforce hierarchy structure
- **Status management** — ticket lifecycle states (open, in-progress, done)
- **`slope next [<archetype>]`** — deterministic selection of the next actionable ticket, based on priority, status, or other sortable fields
- **Metadata from cli** - support to inject metadata from cli (example ```slope <command> --meta-key1=value1 --meta-key2=value2 ...```)
- **`slope init`** — project initialization with optionally bundled default archetypes
- **`slope run <id> -- <command>`** — execute a command with `SLOPE_SCOPE_ID` set, enabling agent workflows like `slope run todo-ccc -- agent --yolo`
