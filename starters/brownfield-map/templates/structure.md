# {{ .StageName }} - {{ .ProjectName }}

## Context
{{- range .Inputs }}
<artifact name="{{ .Name }}">
{{ .Content }}
</artifact>

{{- end }}

You will also receive repository files in `<file path="...">...</file>` blocks.

## Task
Produce `planning/codebase/STRUCTURE.md`.

## Output Requirements
Document:
- Top-level directories and their purpose
- Where core business logic lives
- Where APIs/routes/controllers live
- Where data access lives
- Where UI/views/components live (if applicable)
- Where config lives
- Where tests live

## Output Format
Markdown with these exact headers:
- `# Structure`
- `## Top-Level Map`
- `## Key Subsystems`
- `## Data Flow (High Level)`
- `## Where To Make Changes`
- `## Gaps / Uncertainties`

## Output Format Constraints
CRITICAL: Output ONLY the raw markdown content for `planning/codebase/STRUCTURE.md`.
Do not include conversational text or code fences.
