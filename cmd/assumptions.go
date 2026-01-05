package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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

		prompt := generateAssumptionsPrompt(string(content), specPath)

		prompt = applyMaxChars(prompt, stageMaxChars)
		formatted, err := formatPrompt(stageFormat, "assumptions", prompt)
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

func generateAssumptionsPrompt(content, path string) string {
	return fmt.Sprintf(`List all implicit assumptions in this specification.

For each assumption identified:

### 1. State the Assumption
Clearly articulate what is being assumed but not explicitly stated.

### 2. Why It Matters
Explain the consequence if this assumption is relied upon.

### 3. What Breaks If It's Wrong
Describe the failure modes if this assumption proves false:
- Technical failures
- User experience degradation
- Security vulnerabilities
- Performance issues

### 4. How to Validate It
Provide concrete steps to verify or falsify the assumption:
- Questions to ask stakeholders
- Tests to write
- Metrics to observe
- Prototypes to build

---

## Categories to Consider

- **Environmental assumptions** (OS, runtime, network, hardware)
- **User behavior assumptions** (expertise, intent, workflow)
- **Data assumptions** (volume, format, quality, freshness)
- **Integration assumptions** (APIs, dependencies, versions)
- **Performance assumptions** (latency, throughput, scale)
- **Security assumptions** (threat model, trust boundaries)
- **Temporal assumptions** (timing, ordering, concurrency)

---

## Specification
**Source**: %s

%s
`, path, content)
}
