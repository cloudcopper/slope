// cmd/bootstrap/main.go — Slope CLI MVP (throwaway bootstrap for dogfooding)
// Run via: go run cmd/bootstrap/main.go <command> [flags] [args]
// Stdlib only. Commands: add, prompt.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func main() {
	if len(os.Args) < 2 {
		die("usage: bootstrap <command> [flags] [args]\nCommands: add, prompt")
	}
	switch os.Args[1] {
	case "add":
		cmdAdd(os.Args[2:])
	case "prompt":
		cmdPrompt(os.Args[2:])
	default:
		die("unknown command: " + os.Args[1])
	}
}

func die(msg string) { fmt.Fprintln(os.Stderr, msg); os.Exit(1) }

// findSlope walks up from cwd until it finds a .slope/ directory.
func findSlope() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(cwd, ".slope")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			return "", fmt.Errorf(".slope directory not found")
		}
		cwd = parent
	}
}

// slugify converts a string to a lowercase hyphen-separated slug.
// Only keeps letters, digits, and hyphens; collapses runs into single hyphens.
func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevHyphen := true // suppress leading hyphen
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevHyphen = false
		} else if !prevHyphen {
			b.WriteRune('-')
			prevHyphen = true
		}
	}
	return strings.TrimRight(b.String(), "-")
}

// extractSlug derives a slug from ticket text.
// Uses the first # heading; falls back to the first sentence/line.
func extractSlug(text string) string {
	for _, line := range strings.Split(text, "\n") {
		if trimmed := strings.TrimSpace(line); strings.HasPrefix(trimmed, "#") {
			return slugify(strings.TrimLeft(trimmed, "#"))
		}
	}
	s := text
	if idx := strings.IndexAny(s, ".\n"); idx != -1 {
		s = s[:idx]
	}
	return slugify(s)
}

// ── Frontmatter helpers (no regex) ────────────────────────────────────────────

// fmDelim returns "+++", "---", or "" based on the opening line.
func fmDelim(content string) string {
	switch {
	case strings.HasPrefix(content, "+++"):
		return "+++"
	case strings.HasPrefix(content, "---"):
		return "---"
	}
	return ""
}

// stripFrontmatter removes a TOML (+++) or YAML (---) frontmatter block.
func stripFrontmatter(content string) string {
	delim := fmDelim(content)
	if delim == "" {
		return content
	}
	lines := strings.Split(content, "\n")
	for i := 1; i < len(lines); i++ {
		if strings.TrimRight(lines[i], "\r") == delim {
			return strings.Join(lines[i+1:], "\n")
		}
	}
	return content // no closing delimiter — return as-is
}

// fmGet reads the value of key from TOML or YAML frontmatter.
// Strips surrounding single or double quotes from the value.
func fmGet(content, key string) string {
	delim := fmDelim(content)
	if delim == "" {
		return ""
	}
	lines := strings.Split(content, "\n")
	for i := 1; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], "\r")
		if line == delim {
			break
		}
		for _, sep := range []string{" = ", ": "} {
			if idx := strings.Index(line, sep); idx != -1 {
				if strings.TrimSpace(line[:idx]) == key {
					v := strings.TrimSpace(line[idx+len(sep):])
					if len(v) >= 2 && (v[0] == '"' || v[0] == '\'') {
						v = v[1 : len(v)-1]
					}
					return v
				}
			}
		}
	}
	return ""
}

// fmAddID injects _id into an existing frontmatter block, or creates one.
// No-ops if _id is already present.
func fmAddID(content, id string) string {
	delim := fmDelim(content)
	keySep := " = "
	if delim == "" {
		return "+++\n_id = " + id + "\n+++\n" + content
	}
	if delim == "---" {
		keySep = ": "
	}
	lines := strings.Split(content, "\n")
	for i := 1; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], "\r")
		if line == delim {
			break
		}
		if strings.HasPrefix(line, "_id") {
			return content // already present
		}
	}
	// Insert after the opening delimiter line
	out := make([]string, 0, len(lines)+1)
	out = append(out, lines[0], "_id"+keySep+id)
	out = append(out, lines[1:]...)
	return strings.Join(out, "\n")
}

// ── Ticket operations ─────────────────────────────────────────────────────────

func fileExists(path string) bool { _, err := os.Stat(path); return err == nil }

// findTicket does a depth-first walk of slopeDir for a ticket matching id.
// For README.md files: matches _id frontmatter, or the directory name when _id absent.
// For leaf .md files: matches the filename stem.
// Returns an error if no match or multiple matches are found.
func findTicket(slopeDir, id string) (string, error) {
	var matches []string
	_ = filepath.Walk(slopeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		if filepath.Base(path) == "README.md" {
			data, _ := os.ReadFile(path)
			fmID := fmGet(string(data), "_id")
			dir := filepath.Base(filepath.Dir(path))
			if fmID == id || (fmID == "" && dir == id) {
				matches = append(matches, path)
			}
		} else if strings.TrimSuffix(filepath.Base(path), ".md") == id {
			matches = append(matches, path)
		}
		return nil
	})
	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return "", fmt.Errorf("ticket %q not found", id)
	default:
		return "", fmt.Errorf("multiple tickets match %q", id)
	}
}

// buildPrompt concatenates the bodies (frontmatter stripped) of the ancestor
// chain from the .slope/ root down to the target ticket.
func buildPrompt(slopeDir, ticketPath string) (string, error) {
	rel, err := filepath.Rel(slopeDir, ticketPath)
	if err != nil {
		return "", err
	}
	parts := strings.Split(filepath.ToSlash(rel), "/")

	// Ancestor directories: all path segments except the ticket itself.
	// For a directory ticket (README.md) the containing dir is the ticket —
	// exclude it from ancestors too.
	ancestorParts := parts[:len(parts)-1]
	if parts[len(parts)-1] == "README.md" && len(ancestorParts) > 0 {
		ancestorParts = ancestorParts[:len(ancestorParts)-1]
	}

	var files []string
	cur := slopeDir
	for _, part := range ancestorParts {
		cur = filepath.Join(cur, part)
		if readme := filepath.Join(cur, "README.md"); fileExists(readme) {
			files = append(files, readme)
		}
	}
	files = append(files, ticketPath)

	var sb strings.Builder
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return "", fmt.Errorf("reading %s: %w", f, err)
		}
		body := strings.TrimLeft(stripFrontmatter(string(data)), "\n")
		sb.WriteString(body)
		if len(body) > 0 && !strings.HasSuffix(body, "\n") {
			sb.WriteByte('\n')
		}
	}
	return sb.String(), nil
}

// promoteLeafToDir converts <name>.md → <name>/README.md.
// Injects _id = <name> into frontmatter unless already present.
func promoteLeafToDir(leafPath string) error {
	data, err := os.ReadFile(leafPath)
	if err != nil {
		return err
	}
	id := strings.TrimSuffix(filepath.Base(leafPath), ".md")
	dirPath := strings.TrimSuffix(leafPath, ".md")
	content := fmAddID(string(data), id)
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dirPath, "README.md"), []byte(content), 0o644); err != nil {
		return err
	}
	return os.Remove(leafPath)
}

// resolveParentDir returns the directory where children of parentID are placed.
// Promotes the parent ticket to directory form if it is currently a leaf.
func resolveParentDir(slopeDir, parentID string) (string, error) {
	path, err := findTicket(slopeDir, parentID)
	if err != nil {
		return "", err
	}
	if filepath.Base(path) != "README.md" {
		dirPath := strings.TrimSuffix(path, ".md")
		if err := promoteLeafToDir(path); err != nil {
			return "", fmt.Errorf("promoting %s: %w", parentID, err)
		}
		return dirPath, nil
	}
	return filepath.Dir(path), nil
}

// ── Commands ──────────────────────────────────────────────────────────────────

func cmdAdd(args []string) {
	var parentID string
	var rest []string
	for _, a := range args {
		if strings.HasPrefix(a, "--parent=") {
			parentID = strings.TrimPrefix(a, "--parent=")
		} else {
			rest = append(rest, a)
		}
	}
	if len(rest) < 2 {
		die("usage: add [--parent=<id>] <archetype> \"<text>\"")
	}
	archetype, text := rest[0], rest[1]

	slopeDir, err := findSlope()
	if err != nil {
		die(err.Error())
	}

	slug := extractSlug(text)
	if slug == "" {
		die("could not generate slug from text")
	}
	// Strip redundant archetype prefix (e.g. archetype=feature, slug=feature-foo → foo)
	if strings.HasPrefix(slug, archetype+"-") {
		slug = strings.TrimPrefix(slug, archetype+"-")
	}
	filename := archetype + "-" + slug + ".md"

	targetDir := slopeDir
	if parentID != "" {
		if targetDir, err = resolveParentDir(slopeDir, parentID); err != nil {
			die(err.Error())
		}
	}

	targetPath := filepath.Join(targetDir, filename)
	if fileExists(targetPath) {
		die("file already exists: " + targetPath)
	}
	if err := os.WriteFile(targetPath, []byte(text), 0o644); err != nil {
		die(err.Error())
	}
	fmt.Println(targetPath)
}

func cmdPrompt(args []string) {
	if len(args) < 1 {
		die("usage: prompt <id>")
	}
	slopeDir, err := findSlope()
	if err != nil {
		die(err.Error())
	}
	ticketPath, err := findTicket(slopeDir, args[0])
	if err != nil {
		die(err.Error())
	}
	output, err := buildPrompt(slopeDir, ticketPath)
	if err != nil {
		die(err.Error())
	}
	fmt.Print(output)
}

// Created: 2026-04-25 09:00 UTC, agent: opencode, model: github-copilot/claude-sonnet-4.6
