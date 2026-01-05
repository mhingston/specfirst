package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var failureCmd = &cobra.Command{
	Use:   "failure-modes <spec-file>",
	Short: "Generate a prompt focused on how the spec could fail",
	Long: `Generate a prompt that enumerates all plausible failure modes of a specification.

This command directly addresses happy-path bias by forcing consideration of:
- Partial failures
- Misuse cases
- Ambiguous interpretations
- Race conditions
- Operational edge cases

The output is a structured prompt suitable for AI assistants or human reviewers.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		specPath := args[0]

		content, err := os.ReadFile(specPath)
		if err != nil {
			return fmt.Errorf("reading spec %s: %w", specPath, err)
		}

		prompt := generateFailurePrompt(string(content), specPath)

		prompt = applyMaxChars(prompt, stageMaxChars)
		formatted, err := formatPrompt(stageFormat, "failure-modes", prompt)
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

func generateFailurePrompt(content, path string) string {
	return fmt.Sprintf(`Enumerate all plausible failure modes of this design.

For each failure mode, provide:
- **Description**: What goes wrong
- **Trigger**: How/when it occurs
- **Likelihood**: High / Medium / Low
- **Impact severity**: Critical / High / Medium / Low
- **Detection**: How would you know it happened?
- **Mitigation**: How to prevent or recover

---

## Failure Categories

### Partial Failures
- What happens when some components work but others don't?
- What degraded states are possible?
- What cascading failures could occur?

### Misuse Cases
- How could users intentionally or accidentally break this?
- What abuse patterns are possible?
- What unintended uses could cause harm?

### Ambiguous Interpretations
- Where could different implementers interpret requirements differently?
- What edge cases have undefined behavior?
- What implicit ordering or timing assumptions exist?

### Race Conditions
- What concurrent operations could conflict?
- What ordering assumptions might be violated?
- What atomicity guarantees are missing?

### Operational Edge Cases
- What happens during deployment/rollback?
- What happens during maintenance windows?
- What happens at system boundaries (startup, shutdown, failover)?

### Resource Exhaustion
- What happens when disk/memory/network is depleted?
- What happens under extreme load?
- What happens with malformed or oversized inputs?

### Dependency Failures
- What happens when external services are unavailable?
- What happens with version mismatches?
- What happens with stale caches or expired tokens?

---

## Specification
**Source**: %s

%s
`, path, content)
}
