package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"specfirst/internal/engine/prompt"
	"specfirst/internal/engine/system"
	"specfirst/internal/repository"
)

var assumptionsCmd = &cobra.Command{
	Use:   "assumptions <spec-file>",
	Short: "Generate a prompt to extract implicit assumptions",
	Long: `Generate a prompt that forces the surfacing of hidden assumptions in a specification.

This command helps identify implicit assumptions that could lead to:
- Misunderstandings between stakeholders
- Incorrect implementations
- Untested edge cases
- Deployment failures

The output is a structured prompt suitable for AI assistants or human reviewers.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		specPath := args[0]

		content, err := os.ReadFile(specPath)
		if err != nil {
			return fmt.Errorf("reading spec %s: %w", specPath, err)
		}

		promptStr, err := system.Render("assumptions-extraction.md", system.SpecData{
			Spec: string(content),
		})
		if err != nil {
			return fmt.Errorf("rendering assumptions prompt: %w", err)
		}

		promptStr = prompt.ApplyMaxChars(promptStr, stageMaxChars)
		formatted, err := prompt.Format(stageFormat, "assumptions", promptStr)
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
