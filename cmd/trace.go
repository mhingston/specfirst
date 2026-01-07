package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"specfirst/internal/engine/prompt"
	"specfirst/internal/engine/system"
	"specfirst/internal/repository"
)

var traceCmd = &cobra.Command{
	Use:   "trace <spec-file> <code-file>",
	Short: "Generate a prompt mapping spec requirements to code",
	Long: `Generate a prompt that traces specification requirements to their code implementation.

This command helps verification by:
- Identifying implemented requirements
- flagging missing implementations
- highlighting logic that deviates from spec
- detecting dead code or undocumented features

The output is a structured prompt suitable for AI assistants or human reviewers.`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		specPath := args[0]
		codePath := args[1]

		specContent, err := os.ReadFile(specPath)
		if err != nil {
			return fmt.Errorf("reading spec %s: %w", specPath, err)
		}

		codeContent, err := os.ReadFile(codePath)
		if err != nil {
			return fmt.Errorf("reading code %s: %w", codePath, err)
		}

		promptStr, err := system.Render("trace.md", system.TraceData{
			Spec: string(specContent),
			Code: string(codeContent),
		})
		if err != nil {
			return fmt.Errorf("rendering trace prompt: %w", err)
		}

		promptStr = prompt.ApplyMaxChars(promptStr, stageMaxChars)
		formatted, err := prompt.Format(stageFormat, "trace", promptStr)
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
