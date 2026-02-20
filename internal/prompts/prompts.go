// Package prompts provides system prompt templates for Altera agents.
//
// Templates are embedded Go files using text/template syntax. Each agent role
// (worker, liaison, resolver) has its own template with variables for agent ID,
// task details, rig name, etc. A developer template provides the root CLAUDE.md
// for humans working on the Altera codebase itself.
package prompts

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

//go:embed templates/*.md
var templateFS embed.FS

// Vars holds the template variables available to all prompt templates.
type Vars struct {
	AgentID         string
	TaskID          string
	TaskTitle       string
	TaskDescription string
	RigName         string
	BranchName      string
}

// templates are parsed once at init time.
var templates *template.Template

func init() {
	var err error
	templates, err = template.ParseFS(templateFS, "templates/*.md")
	if err != nil {
		panic(fmt.Sprintf("prompts: parsing templates: %v", err))
	}
}

// render executes a named template with the given variables.
func render(name string, v Vars) (string, error) {
	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, name, v); err != nil {
		return "", fmt.Errorf("rendering %s: %w", name, err)
	}
	return buf.String(), nil
}

// Worker renders the worker agent system prompt.
func Worker(v Vars) (string, error) {
	return render("worker.md", v)
}

// Liaison renders the liaison agent system prompt.
func Liaison(v Vars) (string, error) {
	return render("liaison.md", v)
}

// Resolver renders the resolver agent system prompt.
func Resolver(v Vars) (string, error) {
	return render("resolver.md", v)
}

// Developer renders the root CLAUDE.md for developers working on Altera itself.
// This template does not use any variables.
func Developer() (string, error) {
	return render("developer.md", Vars{})
}
