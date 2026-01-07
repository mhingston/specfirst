package system

import (
	"bytes"
	"embed"
	"fmt"
	"sync"
	"text/template"
)

//go:embed assets/*.md
var promptFS embed.FS

// PromptTemplates holds all parsed prompt templates.
var PromptTemplates *template.Template
var loadOnce sync.Once
var loadErr error

// Load loads the prompts from the embedded filesystem.
// Sev2 Fix: Removed panic in init(). Now also idempotent.
func Load() error {
	loadOnce.Do(func() {
		var err error
		// Parse prompt-contract.md first
		PromptTemplates, err = template.New("prompts").ParseFS(promptFS, "assets/prompt-contract.md")
		if err != nil {
			loadErr = fmt.Errorf("failed to parse prompt-contract.md: %w", err)
			return
		}
		// Parse all other templates
		PromptTemplates, err = PromptTemplates.ParseFS(promptFS, "assets/*.md")
		if err != nil {
			loadErr = fmt.Errorf("failed to parse prompt templates: %w", err)
			return
		}
	})
	return loadErr
}

// SpecData is the data structure for single-spec prompts.
type SpecData struct {
	Spec   string
	Source string
}

// DiffData is the data structure for diff/comparison prompts.
type DiffData struct {
	SpecBefore string
	SpecAfter  string
}

// TraceData is the data structure for trace prompts.
type TraceData struct {
	Spec string
	Code string
}

// Render executes a named prompt template with the given data.
func Render(name string, data interface{}) (string, error) {
	if err := Load(); err != nil {
		return "", err
	}
	var buf bytes.Buffer
	// The template name will be the filename, e.g. "trace.md".
	// ParseFS uses base names for template names.
	if err := PromptTemplates.ExecuteTemplate(&buf, name, data); err != nil {
		return "", fmt.Errorf("rendering prompt %s: %w", name, err)
	}
	return buf.String(), nil
}
