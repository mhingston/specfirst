package prompt

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"specfirst/internal/domain"
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

// GenerateHeader creates a YAML metadata header from a PromptConfig.
func GenerateHeader(config *domain.PromptConfig) string {
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
func ExtractHeader(prompt string) (header string, body string) {
	trimmed := strings.TrimSpace(prompt)
	if !strings.HasPrefix(trimmed, "---") {
		return "", prompt
	}

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
