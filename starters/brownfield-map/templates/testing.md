# {{ .StageName }} - {{ .ProjectName }}

## Context
{{- range .Inputs }}
<artifact name="{{ .Name }}">
{{ .Content }}
</artifact>

{{- end }}

You will also receive repository files in `<file path="...">...</file>` blocks.

## Task
Produce `planning/codebase/TESTING.md`.

## Output Requirements
Document:
- Test frameworks and runners
- Unit vs integration vs e2e patterns
- How to run tests locally
- Common fixtures/helpers
- Any flaky/slow suites (if visible)

## Output Format
Markdown with these exact headers:
- `# Testing`
- `## Tooling`
- `## Test Types`
- `## How To Run`
- `## Patterns & Helpers`
- `## Known Pain Points`
- `## Gaps / Uncertainties`

## Output Format Constraints
CRITICAL: Output ONLY the raw markdown content for `planning/codebase/TESTING.md`.
Do not include conversational text or code fences.
