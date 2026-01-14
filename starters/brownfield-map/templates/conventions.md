# {{ .StageName }} - {{ .ProjectName }}

## Context
{{- range .Inputs }}
<artifact name="{{ .Name }}">
{{ .Content }}
</artifact>

{{- end }}

You will also receive repository files in `<file path="...">...</file>` blocks.

## Task
Produce `planning/codebase/CONVENTIONS.md`.

## Output Requirements
Capture conventions that future changes should follow:
- Naming conventions (files, folders, symbols)
- Layering boundaries and import rules (if evident)
- Error patterns and logging patterns
- How config is loaded
- How dependency injection/service wiring works (if any)
- Formatting/linting rules

## Output Format
Markdown with these exact headers:
- `# Conventions`
- `## Naming`
- `## Project Layout Rules`
- `## Errors & Logging`
- `## Configuration`
- `## Testing Conventions`
- `## Formatting & Linting`
- `## Gaps / Uncertainties`

## Output Format Constraints
CRITICAL: Output ONLY the raw markdown content for `planning/codebase/CONVENTIONS.md`.
Do not include conversational text or code fences.
