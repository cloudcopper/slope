package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"


)

func main() {
	if len(os.Args) < 2 {
		fatalf("usage: slope <command> [flags] [args]\ncommands: add, prompt")
	}
	switch os.Args[1] {
	case "add":
		cmdAdd(os.Args[2:])
	case "prompt":
		cmdPrompt(os.Args[2:])
	default:
		fatalf("unknown command: %s", os.Args[1])
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

// findSlopeRoot walks up from cwd to find .slope/ directory, returns the repo root.
func findSlopeRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		fatalf("getwd: %v", err)
	}
	for {
		if info, err := os.Stat(filepath.Join(dir, ".slope")); err == nil && info.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			fatalf(".slope/ directory not found")
		}
		dir = parent
	}
}

func slopeDir() string {
	return filepath.Join(findSlopeRoot(), ".slope")
}

// --- add command ---

func cmdAdd(args []string) {
	var parentID string
	var positional []string
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "--parent=") {
			parentID = strings.TrimPrefix(args[i], "--parent=")
		} else {
			positional = append(positional, args[i])
		}
	}
	if len(positional) < 2 {
		fatalf("usage: slope add [--parent=<id>] <archetype> \"<text>\"")
	}
	archetype := positional[0]
	bodyText := positional[1]

	slug := generateSlug(bodyText)
	if slug == "" {
		fatalf("could not generate slug from text")
	}
	filename := archetype + "-" + slug + ".md"

	slope := slopeDir()
	targetDir := slope

	if parentID != "" {
		parentPath := findTicket(slope, parentID)
		if parentPath == "" {
			fatalf("parent %q not found", parentID)
		}
		targetDir = promoteIfLeaf(parentPath)
	}

	dest := filepath.Join(targetDir, filename)
	if _, err := os.Stat(dest); err == nil {
		fatalf("file already exists: %s", dest)
	}

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(dest, []byte(bodyText), 0o644); err != nil {
		fatalf("write: %v", err)
	}
	fmt.Println(dest)
}

func generateSlug(body string) string {
	// Try first # heading
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			text := strings.TrimLeft(trimmed, "# ")
			return slugify(text)
		}
	}
	// Fallback: first sentence
	sentence := body
	if idx := strings.IndexAny(body, ".\n"); idx > 0 {
		sentence = body[:idx]
	}
	return slugify(sentence)
}

var nonAlnumHyphen = regexp.MustCompile(`[^a-z0-9-]+`)
var multiHyphen = regexp.MustCompile(`-{2,}`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return '-'
	}, s)
	s = nonAlnumHyphen.ReplaceAllString(s, "")
	s = multiHyphen.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// promoteIfLeaf converts a leaf ticket file into directory form, returns the directory path.
// Promotion rule: <archetype>-<slug>.md → <archetype>-<slug>/README.md
// On promotion, _id is added to frontmatter (unless already present).
func promoteIfLeaf(ticketPath string) string {
	info, err := os.Stat(ticketPath)
	if err != nil {
		fatalf("stat: %v", err)
	}
	if info.IsDir() {
		// Already a directory ticket; README.md is inside
		return ticketPath
	}
	// It's a file — promote to directory
	dir := strings.TrimSuffix(ticketPath, ".md")
	ticketID := strings.TrimSuffix(filepath.Base(ticketPath), ".md")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fatalf("mkdir: %v", err)
	}

	// Read existing content and inject _id if missing
	data, err := os.ReadFile(ticketPath)
	if err != nil {
		fatalf("read: %v", err)
	}
	data = injectIDFrontmatter(data, ticketID)

	newPath := filepath.Join(dir, "README.md")
	if err := os.WriteFile(newPath, data, 0o644); err != nil {
		fatalf("write: %v", err)
	}
	if err := os.Remove(ticketPath); err != nil {
		fatalf("remove: %v", err)
	}
	return dir
}

// injectIDFrontmatter adds _id to TOML frontmatter if not already present.
// If no frontmatter exists, it prepends a new block.
func injectIDFrontmatter(data []byte, id string) []byte {
	content := string(data)

	// Check if frontmatter already contains _id
	if strings.Contains(content, "_id") {
		return data
	}

	idLine := fmt.Sprintf("_id = %q", id)

	// Has TOML frontmatter (+++)?
	if strings.HasPrefix(content, "+++\n") {
		// Insert _id after opening +++
		return []byte("+++\n" + idLine + "\n" + content[4:])
	}

	// Has YAML frontmatter (---)?
	if strings.HasPrefix(content, "---\n") {
		return []byte("---\n" + idLine + "\n" + content[4:])
	}

	// No frontmatter — prepend TOML block
	return []byte("+++\n" + idLine + "\n+++\n\n" + content)
}

// --- find ticket by ID (BFS) ---

// findTicket does BFS of slopeRoot to find a ticket matching id.
// Returns the full path to the ticket file (.md) or directory.
func findTicket(slopeRoot, id string) string {
	var matches []string
	queue := []string{slopeRoot}

	for len(queue) > 0 {
		dir := queue[0]
		queue = queue[1:]
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			name := e.Name()
			full := filepath.Join(dir, name)
			if name == "archetype" {
				continue
			}
			if e.IsDir() {
				// Check if directory ticket matches (README.md with _id)
				stem := name
				if stem == id {
					matches = append(matches, full)
				} else {
					// Check README.md inside for _id
					readmePath := filepath.Join(full, "README.md")
					if fid := readFrontmatterID(readmePath); fid == id {
						matches = append(matches, full)
					}
				}
				queue = append(queue, full)
			} else if strings.HasSuffix(name, ".md") && name != "README.md" {
				stem := strings.TrimSuffix(name, ".md")
				if stem == id {
					matches = append(matches, full)
				}
			}
		}
	}

	switch len(matches) {
	case 0:
		return ""
	case 1:
		return matches[0]
	default:
		fatalf("multiple matches for %q: %v", id, matches)
		return ""
	}
}

// readFrontmatterID reads _id from a file's frontmatter. Returns "" if not found.
func readFrontmatterID(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	content := string(data)

	var body string
	if strings.HasPrefix(content, "+++\n") {
		if end := strings.Index(content[4:], "\n+++"); end >= 0 {
			body = content[4 : 4+end]
		}
	} else if strings.HasPrefix(content, "---\n") {
		if end := strings.Index(content[4:], "\n---"); end >= 0 {
			body = content[4 : 4+end]
		}
	}
	if body == "" {
		return ""
	}

	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		// TOML: _id = "value"
		if strings.HasPrefix(line, "_id") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				v := strings.TrimSpace(parts[1])
				v = strings.Trim(v, "\"'")
				return v
			}
		}
		// YAML: _id: value
		if strings.HasPrefix(line, "_id:") {
			v := strings.TrimSpace(strings.TrimPrefix(line, "_id:"))
			v = strings.Trim(v, "\"'")
			return v
		}
	}
	return ""
}

// findTicketFile returns the .md file path for a ticket (resolving directory tickets).
func findTicketFile(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		fatalf("stat: %v", err)
	}
	if !info.IsDir() {
		return path
	}
	// Directory ticket: README.md inside
	mdFile := filepath.Join(path, "README.md")
	if _, err := os.Stat(mdFile); err == nil {
		return mdFile
	}
	fatalf("ticket directory %s has no README.md file", path)
	return ""
}

// --- prompt command ---

func cmdPrompt(args []string) {
	if len(args) != 1 {
		fatalf("usage: slope prompt <id>")
	}
	id := args[0]
	slope := slopeDir()
	ticketPath := findTicket(slope, id)
	if ticketPath == "" {
		fatalf("ticket %q not found", id)
	}

	ticketFile := findTicketFile(ticketPath)

	// Build ancestor chain by walking from .slope/ to the ticket
	chain := buildAncestorChain(slope, ticketFile)

	// Build merge sequence: for each level, archetype then ticket
	var files []string
	for _, tf := range chain {
		archetype := extractArchetype(tf)
		archetypeFile := resolveArchetype(slope, archetype)
		if archetypeFile != "" {
			files = append(files, archetypeFile)
		}
		files = append(files, tf)
	}

	// Merge all files
	merged := mergeFiles(files)

	// Render to markdown
	fmt.Print(renderAST(merged))
}

// buildAncestorChain returns list of .md files from root to target, inclusive.
func buildAncestorChain(slopeRoot, ticketFile string) []string {
	// Get relative path from .slope/ to the ticket file
	rel, err := filepath.Rel(slopeRoot, ticketFile)
	if err != nil {
		fatalf("rel: %v", err)
	}

	parts := strings.Split(rel, string(filepath.Separator))
	// parts like: ["story-aaa", "feature-bbb", "README.md"]
	// Each directory level that has a README.md is an ancestor
	var chain []string
	for i := 0; i < len(parts)-1; i++ {
		dirPath := filepath.Join(slopeRoot, filepath.Join(parts[:i+1]...))
		readme := filepath.Join(dirPath, "README.md")
		if _, err := os.Stat(readme); err == nil {
			chain = append(chain, readme)
		}
	}
	// Add the ticket file itself (if not already the last entry)
	if len(chain) == 0 || chain[len(chain)-1] != ticketFile {
		chain = append(chain, ticketFile)
	}
	return chain
}

// extractArchetype derives the archetype name from a ticket file path.
// For README.md files, it uses the parent directory name or _id frontmatter.
// For regular files like feature-xxx.md, it uses the filename stem prefix.
func extractArchetype(ticketFile string) string {
	base := filepath.Base(ticketFile)
	if base == "README.md" {
		// Use _id from frontmatter, or fall back to parent dir name
		if fid := readFrontmatterID(ticketFile); fid != "" {
			if idx := strings.Index(fid, "-"); idx > 0 {
				return fid[:idx]
			}
		}
		dirName := filepath.Base(filepath.Dir(ticketFile))
		if idx := strings.Index(dirName, "-"); idx > 0 {
			return dirName[:idx]
		}
		return dirName
	}
	stem := strings.TrimSuffix(base, ".md")
	if idx := strings.Index(stem, "-"); idx > 0 {
		return stem[:idx]
	}
	return stem
}

func resolveArchetype(slopeRoot, archetype string) string {
	// Project-local
	p := filepath.Join(slopeRoot, "archetype", archetype+".md")
	if _, err := os.Stat(p); err == nil {
		return p
	}
	// User-global
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	p = filepath.Join(home, ".config", "slope", "archetype", archetype+".md")
	if _, err := os.Stat(p); err == nil {
		return p
	}
	return ""
}

// --- AST merge ---

// chapter represents a heading and its body content.
type chapter struct {
	level   int    // 0 = preamble (before any heading)
	heading string // heading text (empty for preamble)
	raw     []byte // raw markdown content (heading line + body)
}

var headingRe = regexp.MustCompile(`^(#{1,6})\s+(.*)$`)

// parseChapters splits markdown source into chapters using line-based heading detection.
// We still use goldmark for the final render, but for chapter splitting raw lines are reliable.
func parseChapters(source []byte) []chapter {
	lines := strings.Split(string(source), "\n")

	type headingInfo struct {
		level    int
		text     string
		lineIdx  int
	}

	var headings []headingInfo
	for i, line := range lines {
		if m := headingRe.FindStringSubmatch(line); m != nil {
			headings = append(headings, headingInfo{
				level:   len(m[1]),
				text:    strings.TrimSpace(m[2]),
				lineIdx: i,
			})
		}
	}

	if len(headings) == 0 {
		return []chapter{{level: 0, raw: source}}
	}

	var chapters []chapter

	// Preamble before first heading
	if headings[0].lineIdx > 0 {
		preamble := strings.Join(lines[:headings[0].lineIdx], "\n")
		preamble = strings.TrimRight(preamble, "\n")
		if len(strings.TrimSpace(preamble)) > 0 {
			chapters = append(chapters, chapter{level: 0, raw: []byte(preamble + "\n")})
		}
	}

	for i, h := range headings {
		endLine := len(lines)
		if i+1 < len(headings) {
			endLine = headings[i+1].lineIdx
		}
		raw := strings.Join(lines[h.lineIdx:endLine], "\n")
		chapters = append(chapters, chapter{
			level:   h.level,
			heading: h.text,
			raw:     []byte(raw),
		})
	}

	return chapters
}

// extractBody returns the body of a chapter (everything after the heading line).
func extractBody(ch chapter) []byte {
	if ch.level == 0 {
		return ch.raw
	}
	// Skip the first line (the heading)
	idx := bytes.IndexByte(ch.raw, '\n')
	if idx < 0 {
		return nil
	}
	return bytes.TrimSpace(ch.raw[idx+1:])
}

// mergeFiles performs heading-based merge across multiple files.
func mergeFiles(files []string) []chapter {
	var merged []chapter

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			fatalf("read %s: %v", f, err)
		}
		chapters := parseChapters(data)
		merged = mergeChapters(merged, chapters)
	}

	return merged
}

func mergeChapters(base, overlay []chapter) []chapter {
	result := make([]chapter, len(base))
	copy(result, base)

	for _, ch := range overlay {
		if ch.level == 0 {
			// Preamble: append to existing preamble or insert after last preamble
			found := false
			for i, r := range result {
				if r.level == 0 {
					combined := bytes.TrimRight(r.raw, "\n")
					combined = append(combined, '\n')
					combined = append(combined, bytes.TrimRight(ch.raw, "\n")...)
					combined = append(combined, '\n')
					result[i] = chapter{level: 0, raw: combined}
					found = true
					break
				}
			}
			if !found {
				result = append(result, ch)
			}
			continue
		}

		// Find existing chapter with same heading text
		found := false
		for i, r := range result {
			if r.heading == ch.heading {
				body := extractBody(ch)
				if len(body) == 0 {
					// Empty body: remove chapter
					result = append(result[:i], result[i+1:]...)
				} else {
					// Replace
					result[i] = ch
				}
				found = true
				break
			}
		}
		if !found {
			result = append(result, ch)
		}
	}

	return result
}

// renderAST renders merged chapters back to markdown.
func renderAST(chapters []chapter) string {
	var buf bytes.Buffer
	for i, ch := range chapters {
		if i > 0 {
			// Ensure blank line between chapters
			if !bytes.HasSuffix(buf.Bytes(), []byte("\n\n")) {
				if bytes.HasSuffix(buf.Bytes(), []byte("\n")) {
					buf.WriteByte('\n')
				} else {
					buf.WriteString("\n\n")
				}
			}
		}
		buf.Write(bytes.TrimRight(ch.raw, "\n"))
		buf.WriteByte('\n')
	}
	return buf.String()
}
