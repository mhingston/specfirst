package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var testIntentCmd = &cobra.Command{
	Use:   "test-intent <spec-file>",
	Short: "Generate a test intent prompt (not test code)",
	Long: `Generate a prompt that derives comprehensive test intent from a specification.

This command bridges specification thinking and implementation safely by producing
test INTENT rather than test CODE. This keeps SpecFirst out of code generation
territory while providing valuable guidance for any language or framework.

The output is a structured prompt suitable for AI assistants or human testers.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		specPath := args[0]

		content, err := os.ReadFile(specPath)
		if err != nil {
			return fmt.Errorf("reading spec %s: %w", specPath, err)
		}

		prompt := generateTestIntentPrompt(string(content), specPath)

		prompt = applyMaxChars(prompt, stageMaxChars)
		formatted, err := formatPrompt(stageFormat, "test-intent", prompt)
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

func generateTestIntentPrompt(content, path string) string {
	return fmt.Sprintf(`Derive a comprehensive test intent from this specification.

Do NOT generate test code. Instead, describe WHAT should be tested and WHY.

---

## Required Invariants
Properties that must ALWAYS hold, regardless of input or state:
- List each invariant as a clear assertion
- Explain what system property it protects
- Note how violations could be detected

## Boundary Conditions
Edge cases at the limits of valid behavior:
- Minimum and maximum values
- Empty, null, or missing inputs
- First and last items in sequences
- Transitions between states
- Timeout and retry boundaries

## Negative Cases
What should NOT happen under any circumstances:
- Forbidden state transitions
- Security violations
- Data corruption scenarios
- Resource leaks
- Incorrect error suppression

## Happy Path Scenarios
The expected successful flows:
- Primary use cases with typical inputs
- Expected state changes
- Correct output formats

## Error Handling
How errors should be handled:
- Expected error conditions
- Error message quality
- Recovery behavior
- Error propagation patterns

## Performance Expectations
Non-functional requirements to verify:
- Response time bounds
- Throughput requirements
- Resource usage limits
- Scalability expectations

## Observability Requirements
What needs to be visible for debugging and monitoring:
- Metrics that should be exposed
- Logs that should be generated
- Traces that should be emitted
- Alerts that should fire on failure

## Integration Points
Where this component interacts with others:
- API contract compliance
- Data format compatibility
- Sequence diagram coverage
- Failure injection points

---

## Specification
**Source**: %s

%s
`, path, content)
}
