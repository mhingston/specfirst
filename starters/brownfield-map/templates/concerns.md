# {{ .StageName }} - {{ .ProjectName }}

## Context
{{- range .Inputs }}
<artifact name="{{ .Name }}">
{{ .Content }}
</artifact>

{{- end }}

You will also receive repository files in `<file path="...">...</file>` blocks.

## Task
Produce `planning/codebase/CONCERNS.md`.

## Output Requirements
Identify:
- Fragile areas and tech debt hotspots
- Security risks (secrets, auth, input handling)
- Performance risks (N+1, missing indexes, excessive IO)
- Reliability risks (error handling gaps, missing retries, lack of idempotency)
- Maintainability risks (tight coupling, unclear boundaries)

For each concern, include:
- Evidence (what file(s) suggest it)
- Impact
- Suggested mitigation direction (high level, not a refactor plan)

## Output Format
Markdown with these exact headers:
- `# Concerns`
- `## High-Risk Areas`
- `## Security`
- `## Performance`
- `## Reliability`
- `## Maintainability`
- `## Recommended Next Checks`

## Output Format Constraints
CRITICAL: Output ONLY the raw markdown content for `planning/codebase/CONCERNS.md`.
Do not include conversational text or code fences.
