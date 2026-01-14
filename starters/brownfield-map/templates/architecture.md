# {{ .StageName }} - {{ .ProjectName }}

## Context
{{- range .Inputs }}
<artifact name="{{ .Name }}">
{{ .Content }}
</artifact>

{{- end }}

You will also receive repository files in `<file path="...">...</file>` blocks.

## Task
Produce `planning/codebase/ARCHITECTURE.md`.

## Output Requirements
Describe:
- Major modules/services and how they interact
- Key data models/entities (high level)
- Request flow for the primary user path(s)
- Error handling strategy (where failures surface)
- Any async/background processing

## Output Format
Markdown with these exact headers:
- `# Architecture`
- `## High-Level Overview`
- `## Key Components`
- `## Request / Execution Flow`
- `## Data Model (Conceptual)`
- `## Failure Modes & Error Handling`
- `## Performance & Scaling Notes`
- `## Gaps / Uncertainties`

## Rules
- Stay descriptive, not prescriptive.
- If multiple architectures exist, call out boundaries.

## Output Format Constraints
CRITICAL: Output ONLY the raw markdown content for `planning/codebase/ARCHITECTURE.md`.
Do not include conversational text or code fences.
