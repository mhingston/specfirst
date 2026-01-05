package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var traceCmd = &cobra.Command{
	Use:   "trace <spec-file>",
	Short: "Generate a prompt for spec-to-code mapping",
	Long: `Generate a prompt that asks for mapping between specification sections and code areas.

This command helps identify:
- Which code modules implement which spec sections
- Missing implementation coverage
- Dead or obsolete code risks
- Refactoring impact areas

The output is a structured prompt suitable for AI assistants or human reviewers.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		specPath := args[0]

		content, err := os.ReadFile(specPath)
		if err != nil {
			return fmt.Errorf("reading spec %s: %w", specPath, err)
		}

		prompt := generateTracePrompt(string(content), specPath)

		prompt = applyMaxChars(prompt, stageMaxChars)
		formatted, err := formatPrompt(stageFormat, "trace", prompt)
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

func generateTracePrompt(content, path string) string {
	return fmt.Sprintf(`For each section of this specification, create a traceability mapping.

---

## For Each Spec Section

### 1. Identify Code Modules Affected
- Which files, packages, or modules implement this section?
- Which functions or classes are directly involved?
- Which configuration or infrastructure is required?

### 2. Note Implementation Coverage
- Is this section fully implemented, partially implemented, or not started?
- What aspects are implemented vs. planned?
- Are there any commented-out or feature-flagged implementations?

### 3. Identify Missing Coverage
- What spec requirements have no corresponding code?
- What implicit requirements are not implemented?
- What error cases are not handled?

### 4. Identify Dead or Obsolete Code Risks
- What code exists that no longer maps to current spec?
- What code was built for removed requirements?
- What technical debt is linked to spec changes?

### 5. Assess Change Impact
If this spec section changes, what would be affected?
- Direct code changes required
- Dependent components that would need updates
- Tests that would need modification
- Documentation that would become stale

---

## Output Format

Provide a table or structured list mapping each spec section to:
| Spec Section | Primary Code Location | Coverage Status | Notes |
|--------------|----------------------|-----------------|-------|
| ...          | ...                  | ...             | ...   |

Also flag:
- 🔴 **Not Implemented**: Spec exists but no code
- 🟡 **Partially Implemented**: Some coverage gaps
- 🟢 **Fully Implemented**: Complete coverage
- ⚫ **Obsolete Code**: Code exists but spec removed

---

## Specification
**Source**: %s

%s
`, path, content)
}
