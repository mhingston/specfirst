package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff <old-spec> <new-spec>",
	Short: "Generate a prompt analyzing spec changes",
	Long: `Generate a prompt that describes the delta between two specification files.

This command helps evaluate behavioral differences, backward compatibility risks,
required code changes, and tests that must be updated when specifications change.

The output is a structured prompt suitable for AI assistants or human reviewers.`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		oldPath := args[0]
		newPath := args[1]

		oldContent, err := os.ReadFile(oldPath)
		if err != nil {
			return fmt.Errorf("reading old spec %s: %w", oldPath, err)
		}

		newContent, err := os.ReadFile(newPath)
		if err != nil {
			return fmt.Errorf("reading new spec %s: %w", newPath, err)
		}

		prompt := generateDiffPrompt(string(oldContent), string(newContent), oldPath, newPath)

		prompt = applyMaxChars(prompt, stageMaxChars)
		formatted, err := formatPrompt(stageFormat, "diff", prompt)
		if err != nil {
			return err
		}

		if stageOut != "" {
			if err := writeOutput(stageOut, formatted); err != nil {
				return err
			}
		}
		_, err = cmd.OutOrStdout().Write([]byte(formatted))
		return err
	},
}

func generateDiffPrompt(oldContent, newContent, oldPath, newPath string) string {
	return fmt.Sprintf(`Given the following previous specification and proposed changes, evaluate:

1. **Behavioral differences** — What has changed in expected system behavior?
2. **Backward compatibility risks** — What existing functionality might break?
3. **Required code changes** — What implementation updates are needed?
4. **Tests that must be updated or added** — What test coverage is affected?

For each identified change:
- Describe the nature of the change (addition, removal, modification)
- Assess the impact scope (isolated, cross-cutting, breaking)
- Suggest migration steps if applicable

---

## Previous Specification
**Source**: %s

%s

---

## Proposed Specification
**Source**: %s

%s
`, oldPath, oldContent, newPath, newContent)
}
