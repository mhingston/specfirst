package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"specfirst/internal/engine/prompt"
	"specfirst/internal/repository"
)

var reviewPersona string

var reviewCmd = &cobra.Command{
	Use:   "review <spec-file>",
	Short: "Generate a persona-based review prompt",
	Long: `Generate a role-based review prompt for a specification.

This command creates structured review prompts from different perspectives:
- security: Attack surfaces, data exposure, privilege boundaries
- performance: Bottlenecks, scalability, resource usage
- maintainer: Complexity, coupling, documentation gaps
- accessibility: Inclusivity, standards compliance
- user: UX gaps, usability issues

The output is a structured prompt suitable for AI assistants or human reviewers.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		specPath := args[0]

		content, err := os.ReadFile(specPath)
		if err != nil {
			return fmt.Errorf("reading spec %s: %w", specPath, err)
		}

		promptStr, err := generateReviewPrompt(string(content), specPath, reviewPersona)
		if err != nil {
			return err
		}

		promptStr = prompt.ApplyMaxChars(promptStr, stageMaxChars)
		formatted, err := prompt.Format(stageFormat, "review", promptStr)
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

func init() {
	reviewCmd.Flags().StringVar(&reviewPersona, "persona", "security",
		"Review persona: security, performance, maintainer, accessibility, user")
}

var reviewPersonas = map[string]string{
	"security": `You are a **security reviewer**. Analyze this specification for:

## Attack Surfaces
- What entry points could be exploited?
- What inputs are not validated or sanitized?
- What authentication/authorization gaps exist?

## Data Exposure
- What sensitive data is handled?
- How is data protected at rest and in transit?
- What logging might leak sensitive information?

## Privilege Boundaries
- What trust boundaries are crossed?
- What principle-of-least-privilege violations exist?
- What escalation paths are possible?

## Missing Constraints
- What rate limiting is needed?
- What input validation is missing?
- What error handling could leak information?

For each finding, rate the severity (Critical/High/Medium/Low) and suggest mitigations.`,

	"performance": `You are a **performance reviewer**. Analyze this specification for:

## Bottlenecks
- What operations could become slow at scale?
- What blocking or synchronous operations exist?
- What network round-trips are involved?

## Scalability
- How does this design scale horizontally?
- What are the limiting factors?
- What shared resources could create contention?

## Resource Usage
- What memory patterns are concerning?
- What CPU-intensive operations exist?
- What I/O patterns could cause issues?

## Latency & Throughput
- What are the expected latency requirements?
- What queuing behavior is expected?
- What caching opportunities exist?

For each finding, estimate the impact and suggest optimizations.`,

	"maintainer": `You are a **maintainability reviewer**. Analyze this specification for:

## Complexity
- What areas are overly complex?
- What could be simplified without losing functionality?
- What indirection adds cognitive overhead?

## Coupling
- What tight coupling between components exists?
- What changes would cascade across the system?
- What dependencies could be reduced?

## Documentation Gaps
- What behavior is ambiguous or underspecified?
- What edge cases are not addressed?
- What operational knowledge is missing?

## Evolution Risks
- What assumptions will break as the system evolves?
- What extension points are needed but missing?
- What technical debt is being introduced?

For each finding, suggest refactoring or documentation improvements.`,

	"accessibility": `You are an **accessibility reviewer**. Analyze this specification for:

## Inclusivity
- What barriers exist for users with disabilities?
- What assistive technology support is needed?
- What modality assumptions are made (sight, hearing, motor)?

## Standards Compliance
- What WCAG guidelines are relevant?
- What accessibility requirements are missing?
- What testing approaches are needed?

## Universal Design
- What internationalization considerations apply?
- What cultural assumptions are made?
- What reading level is assumed?

## Error Handling
- How are errors communicated accessibly?
- What recovery paths exist for all users?
- What help and documentation is available?

For each finding, reference relevant standards and suggest improvements.`,

	"user": `You are a **user experience reviewer**. Analyze this specification for:

## Usability Gaps
- What user journeys are unclear or cumbersome?
- What mental model mismatches could occur?
- What feedback loops are missing?

## Edge Cases
- What happens when users make mistakes?
- What empty or error states are undefined?
- What first-time user experiences are considered?

## Consistency
- What interaction patterns are inconsistent?
- What terminology is ambiguous?
- What visual hierarchy issues exist?

## User Goals
- What primary user goals are supported?
- What goals are blocked or hindered?
- What task completion signals exist?

For each finding, describe the user impact and suggest improvements.`,
}

func generateReviewPrompt(content, path, persona string) (string, error) {
	personaLower := strings.ToLower(strings.TrimSpace(persona))
	reviewInstructions, ok := reviewPersonas[personaLower]
	if !ok {
		validPersonas := make([]string, 0, len(reviewPersonas))
		for k := range reviewPersonas {
			validPersonas = append(validPersonas, k)
		}
		return "", fmt.Errorf("unknown persona %q; valid options: %s", persona, strings.Join(validPersonas, ", "))
	}

	return fmt.Sprintf(`%s

---

## Specification
**Source**: %s

%s
`, reviewInstructions, path, content), nil
}
