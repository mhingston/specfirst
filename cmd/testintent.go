package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"specfirst/internal/engine/prompt"
	"specfirst/internal/engine/system"
	"specfirst/internal/repository"
)

// GenerateTestIntentPrompt generates a formatted prompt string for testing strategy based on the given spec content.
func GenerateTestIntentPrompt(specContent string) (string, error) {
	promptStr, err := system.Render("test-intent.md", system.SpecData{
		Spec: specContent,
	})
	if err != nil {
		return "", fmt.Errorf("rendering test-intent prompt: %w", err)
	}

	promptStr = prompt.ApplyMaxChars(promptStr, stageMaxChars)
	formatted, err := prompt.Format(stageFormat, "test-intent", promptStr)
	if err != nil {
		return "", err
	}
	return formatted, nil
}

var testIntentCmd = &cobra.Command{
	Use:   "test-intent <spec-file>",
	Short: "Generate a prompt focused on testing strategy",
	Long: `Generate a prompt that derives a comprehensive test strategy from a specification.

This command extracts:
- Functional test cases (happy path & error conditions)
- Integration boundaries
- Property-based testing candidates
- Security test vectors
- Performance benchmarks

The output is a structured prompt suitable for AI assistants or human reviewers.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		specPath := args[0]

		content, err := os.ReadFile(specPath)
		if err != nil {
			return fmt.Errorf("reading spec %s: %w", specPath, err)
		}

		promptStr, err := system.Render("test-intent.md", system.SpecData{
			Spec: string(content),
		})
		if err != nil {
			return fmt.Errorf("rendering test-intent prompt: %w", err)
		}

		promptStr = prompt.ApplyMaxChars(promptStr, stageMaxChars)
		formatted, err := prompt.Format(stageFormat, "test-intent", promptStr)
		if err != nil {
			return err
		}

		if stageOut != "" {
			if err := repository.WriteOutput(stageOut, formatted); err != nil {
				return err
			}
		}
		_, err = cmd.OutOrStdout().Write([]byte(formatted))
		return err
	},
}
