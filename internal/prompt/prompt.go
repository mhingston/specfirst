package prompt

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"specfirst/internal/protocol"
)

// Format formats a prompt string according to the specified format.
func Format(format string, stageID string, prompt string) (string, error) {
	if format == "text" {
		return prompt, nil
	}
	if format == "json" {
		payload := map[string]string{
			"stage":  stageID,
			"prompt": prompt,
		}
		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data) + "\n", nil
	}
	if format == "yaml" {
		// Simple YAML output without external dependencies
		escapedPrompt := strings.ReplaceAll(prompt, "\\", "\\\\")
		escapedPrompt = strings.ReplaceAll(escapedPrompt, "\"", "\\\"")
		return fmt.Sprintf("stage: %s\nprompt: |\n  %s\n", stageID, strings.ReplaceAll(escapedPrompt, "\n", "\n  ")), nil
	}
	if format == "shell" {
		// Escape single quotes for shell safety
		escapedStageID := strings.ReplaceAll(stageID, "'", "'\"'\"'")
		delimiter := heredocDelimiter(prompt)
		return fmt.Sprintf("SPECFIRST_STAGE='%s'\nSPECFIRST_PROMPT=$(cat <<'%s'\n%s\n%s\n)\n", escapedStageID, delimiter, prompt, delimiter), nil
	}
	return "", errors.New("unsupported format: " + format)
}

// ApplyMaxChars truncates a prompt to a maximum number of characters.
func ApplyMaxChars(prompt string, maxChars int) string {
	if maxChars <= 0 {
		return prompt
	}
	// Use runes to avoid truncating mid-UTF8 character
	runes := []rune(prompt)
	if len(runes) <= maxChars {
		return prompt
	}
	return string(runes[:maxChars])
}

func heredocDelimiter(prompt string) string {
	base := "SPECFIRST_EOF"
	delimiter := base
	for i := 0; ; i++ {
		if i > 0 {
			delimiter = fmt.Sprintf("%s_%d", base, i)
		}
		line := delimiter + "\n"
		if !strings.Contains(prompt, "\n"+delimiter+"\n") &&
			!strings.HasPrefix(prompt, line) &&
			!strings.HasSuffix(prompt, "\n"+delimiter) &&
			prompt != delimiter {
			return delimiter
		}
	}
}

// Schema defines validation rules for generated prompts.
type Schema struct {
	RequiredSections []string // e.g., "Context", "Constraints", "Output Requirements"
	ForbiddenPhrases []string // e.g., "make it better", "improve this"
}

// Merge adds rules from a LintConfig to the schema.
func (s *Schema) Merge(cfg *protocol.LintConfig) {
	if cfg == nil {
		return
	}
	s.RequiredSections = append(s.RequiredSections, cfg.RequiredSections...)
	s.ForbiddenPhrases = append(s.ForbiddenPhrases, cfg.ForbiddenPhrases...)
}

// ValidationResult holds lint warnings for a prompt.
type ValidationResult struct {
	Warnings []string
}

// DefaultSchema returns the built-in prompt schema with sensible defaults.
func DefaultSchema() Schema {
	return Schema{
		RequiredSections: []string{
			"Context",
			"Task",
			"Assumptions",
		},
		ForbiddenPhrases: []string{
			"make it better",
			"improve this",
			"fix it",
			"do your best",
			"be creative",
			"use best practices",
			"make it good",
			"enhance this",
			"optimize this",
			"make it perfect",
		},
	}
}

// Validate checks a prompt against the schema and returns validation warnings.
func Validate(prompt string, schema Schema) ValidationResult {
	var warnings []string

	// Check for required sections
	promptLower := strings.ToLower(prompt)
	for _, section := range schema.RequiredSections {
		// Look for section headers (# Section, ## Section, etc.) with flexible whitespace
		// We use a case-insensitive regex that matches at the beginning of a line
		pattern := "(?mi)^#+\\s+" + regexp.QuoteMeta(strings.TrimSpace(section)) + "\\s*$"
		re, err := regexp.Compile(pattern)
		if err != nil {
			// Fallback to simple contains if regex fails (shouldn't happen with QuoteMeta)
			if !strings.Contains(strings.ToLower(prompt), strings.ToLower(section)) {
				warnings = append(warnings, "missing required section: "+section)
			}
			continue
		}

		if !re.MatchString(prompt) {
			warnings = append(warnings, "missing required section: "+section)
		}
	}

	// Check for forbidden phrases
	for _, phrase := range schema.ForbiddenPhrases {
		phraseLower := strings.ToLower(phrase)
		if strings.Contains(promptLower, phraseLower) {
			warnings = append(warnings, "contains ambiguous phrase: \""+phrase+"\"")
		}
	}

	return ValidationResult{Warnings: warnings}
}

// GenerateHeader creates a YAML metadata header from a PromptConfig.
func GenerateHeader(config *protocol.PromptConfig) string {
	if config == nil {
		return ""
	}

	var lines []string
	lines = append(lines, "---")

	if config.Intent != "" {
		lines = append(lines, "intent: "+config.Intent)
	}
	if config.ExpectedOutput != "" {
		lines = append(lines, "expected_output: "+config.ExpectedOutput)
	}
	if config.Determinism != "" {
		lines = append(lines, "determinism: "+config.Determinism)
	}
	if config.AllowedCreativity != "" {
		lines = append(lines, "allowed_creativity: "+config.AllowedCreativity)
	}
	if config.Granularity != "" {
		lines = append(lines, "granularity: "+config.Granularity)
	}

	lines = append(lines, "---")
	return strings.Join(lines, "\n") + "\n"
}

// ExtractHeader extracts the YAML metadata header from a prompt, if present.
// Returns the header content (without delimiters) and the remaining prompt.
func ExtractHeader(prompt string) (header string, body string) {
	trimmed := strings.TrimSpace(prompt)
	if !strings.HasPrefix(trimmed, "---") {
		return "", prompt
	}

	// Find the closing ---
	rest := strings.TrimPrefix(trimmed, "---")
	rest = strings.TrimPrefix(rest, "\n")

	endIdx := strings.Index(rest, "---")
	if endIdx == -1 {
		return "", prompt
	}

	header = strings.TrimSpace(rest[:endIdx])
	body = strings.TrimSpace(rest[endIdx+3:])
	return header, body
}

// ValidateStructure checks if a prompt has proper structure for its stage type.
func ValidateStructure(prompt string, stageType string) ValidationResult {
	var warnings []string

	switch stageType {
	case "decompose":
		// Decompose prompts should mention task structure
		promptLower := strings.ToLower(prompt)
		if !strings.Contains(promptLower, "task") {
			warnings = append(warnings, "decompose prompt should reference task structure")
		}
		if !strings.Contains(promptLower, "output") && !strings.Contains(promptLower, "format") {
			warnings = append(warnings, "decompose prompt should specify output format")
		}
	case "task_prompt":
		// Task prompts should have clear goal and acceptance criteria
		promptLower := strings.ToLower(prompt)
		if !strings.Contains(promptLower, "goal") && !strings.Contains(promptLower, "objective") {
			warnings = append(warnings, "task prompt should include a goal or objective")
		}
		if !strings.Contains(promptLower, "assumption") {
			warnings = append(warnings, "task prompt should include an assumptions section")
		}
	}

	return ValidationResult{Warnings: warnings}
}

// ContainsAmbiguity checks if the prompt contains ambiguous language.
// This is a more thorough check than just forbidden phrases.
func ContainsAmbiguity(prompt string) []string {
	var issues []string

	// Patterns that indicate vague instructions
	vaguePatterns := []regexp.Regexp{
		*regexp.MustCompile(`(?i)\b(maybe|perhaps|possibly|might)\b\s+.*\b(add|include|consider|use)\b`),
		*regexp.MustCompile(`(?i)\b(as needed|if necessary|when appropriate|where applicable|as appropriate)\b`),
		*regexp.MustCompile(`(?i)\b(etc\.?|and so on|and the like)\b`),
		*regexp.MustCompile(`(?i)\b(some|various|various different|multiple)\b\s+(things|stuff|items|parts|features)\b`),
		*regexp.MustCompile(`(?i)\b(in a way that|to be determined|tbd)\b`),
		*regexp.MustCompile(`(?i)\b(ensure|make sure)\b\s+.*\b(good|better|perfect|nice)\b`),
	}

	for _, pattern := range vaguePatterns {
		if pattern.MatchString(prompt) {
			match := pattern.FindString(prompt)
			issues = append(issues, "vague language detected: \""+match+"\"")
		}
	}

	return issues
}
