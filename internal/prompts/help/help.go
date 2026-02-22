// Package help provides embedded help content for Altera agent types.
//
// Help topics are organized by agent type (worker, liaison) and topic name.
// Content is embedded at compile time using embed.FS.
package help

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

//go:embed worker/*.md liaison/*.md
var helpFS embed.FS

// AgentTypes returns the list of valid agent types.
func AgentTypes() []string {
	return []string{"liaison", "worker"}
}

// Topics returns the list of available topics for an agent type.
func Topics(agentType string) ([]string, error) {
	entries, err := fs.ReadDir(helpFS, agentType)
	if err != nil {
		return nil, fmt.Errorf("unknown agent type: %s", agentType)
	}
	var topics []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".md") {
			topics = append(topics, strings.TrimSuffix(name, ".md"))
		}
	}
	sort.Strings(topics)
	return topics, nil
}

// Lookup finds and returns help content for the given agent type and topic path.
func Lookup(agentType string, topicParts ...string) (string, error) {
	if len(topicParts) == 0 {
		return "", fmt.Errorf("no topic specified")
	}
	filePath := path.Join(agentType, strings.Join(topicParts, "/") + ".md")
	data, err := helpFS.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("topic not found: %s %s", agentType, strings.Join(topicParts, " "))
	}
	return string(data), nil
}
