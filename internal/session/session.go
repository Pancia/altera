// Package session provides access to Claude Code JSONL conversation transcripts.
// Claude Code writes a transcript for each session to
// ~/.claude/projects/{path-encoded-cwd}/{session-id}.jsonl.
package session

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ClaudeProjectsDir is the base directory where Claude Code stores project data.
var ClaudeProjectsDir = filepath.Join(os.Getenv("HOME"), ".claude", "projects")

// EncodePath converts a filesystem path to Claude Code's project directory
// format. Slashes are replaced with hyphens and the path starts with a hyphen.
// Example: /Users/foo/project -> -Users-foo-project
func EncodePath(worktreePath string) string {
	// Claude Code uses the absolute path with / replaced by -
	clean := filepath.Clean(worktreePath)
	return strings.ReplaceAll(clean, "/", "-")
}

// TranscriptDir returns the full path to the Claude Code projects directory
// for a given worktree path.
func TranscriptDir(worktreePath string) string {
	return filepath.Join(ClaudeProjectsDir, EncodePath(worktreePath))
}

// FindTranscripts returns all .jsonl files in the given directory, sorted by
// modification time (newest first).
func FindTranscripts(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading transcript dir: %w", err)
	}

	type fileInfo struct {
		path    string
		modTime time.Time
	}
	var files []fileInfo
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".jsonl" {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, fileInfo{
			path:    filepath.Join(dir, e.Name()),
			modTime: info.ModTime(),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.After(files[j].modTime)
	})

	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = f.path
	}
	return paths, nil
}

// TranscriptEntry represents a single line from a Claude Code JSONL transcript.
type TranscriptEntry struct {
	Type      string          `json:"type"`
	Message   json.RawMessage `json:"message,omitempty"`
	Timestamp string          `json:"timestamp,omitempty"`

	// For assistant/user messages.
	Role    string          `json:"role,omitempty"`
	Content json.RawMessage `json:"content,omitempty"`
}

// RenderTranscript reads a JSONL file and produces a human-readable summary.
// It extracts user/assistant messages, tool use, and timestamps.
func RenderTranscript(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("opening transcript: %w", err)
	}
	defer f.Close()

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Transcript: %s\n\n", filepath.Base(path)))

	scanner := bufio.NewScanner(f)
	// Allow large lines (JSONL entries can be big).
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry map[string]any
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}

		renderEntry(&b, entry)
	}

	if err := scanner.Err(); err != nil {
		return b.String(), fmt.Errorf("scanning transcript: %w", err)
	}

	return b.String(), nil
}

// renderEntry formats a single JSONL entry for human display.
func renderEntry(b *strings.Builder, entry map[string]any) {
	entryType, _ := entry["type"].(string)

	switch entryType {
	case "user":
		renderMessage(b, "USER", entry)
	case "assistant":
		renderMessage(b, "ASSISTANT", entry)
	case "result":
		renderResult(b, entry)
	}
}

// renderMessage formats a user or assistant message.
func renderMessage(b *strings.Builder, role string, entry map[string]any) {
	b.WriteString(fmt.Sprintf("--- %s ---\n", role))

	message, ok := entry["message"].(map[string]any)
	if !ok {
		return
	}

	content, ok := message["content"]
	if !ok {
		return
	}

	switch c := content.(type) {
	case string:
		b.WriteString(truncate(c, 500))
		b.WriteByte('\n')
	case []any:
		for _, block := range c {
			blockMap, ok := block.(map[string]any)
			if !ok {
				continue
			}
			blockType, _ := blockMap["type"].(string)
			switch blockType {
			case "text":
				text, _ := blockMap["text"].(string)
				b.WriteString(truncate(text, 500))
				b.WriteByte('\n')
			case "tool_use":
				name, _ := blockMap["name"].(string)
				b.WriteString(fmt.Sprintf("[tool_use: %s]\n", name))
			case "tool_result":
				b.WriteString("[tool_result]\n")
			}
		}
	}
	b.WriteByte('\n')
}

// renderResult formats a result/tool_result entry.
func renderResult(b *strings.Builder, entry map[string]any) {
	// Results are often very verbose; just note them.
	b.WriteString("[result]\n")
}

// truncate shortens a string to maxLen, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
