package prompt

import (
	"regexp"
	"strings"

	"specfirst/internal/domain"
)

// Schema defines validation rules for generated prompts.
type Schema struct {
	RequiredSections []string
	ForbiddenPhrases []string
}

// Merge adds rules from a LintConfig to the schema.
func (s *Schema) Merge(cfg *domain.LintConfig) {
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
		pattern := "(?mi)^#+\\s+" + regexp.QuoteMeta(strings.TrimSpace(section)) + "\\s*$"
		re, err := regexp.Compile(pattern)
		if err != nil {
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

// ValidateStructure checks if a prompt has proper structure for its stage type.
func ValidateStructure(prompt string, stageType string) ValidationResult {
	var warnings []string

	switch stageType {
	case "decompose":
		promptLower := strings.ToLower(prompt)
		if !strings.Contains(promptLower, "task") {
			warnings = append(warnings, "decompose prompt should reference task structure")
		}
		if !strings.Contains(promptLower, "output") && !strings.Contains(promptLower, "format") {
			warnings = append(warnings, "decompose prompt should specify output format")
		}
	case "task_prompt":
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
func ContainsAmbiguity(prompt string) []string {
	var issues []string

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
