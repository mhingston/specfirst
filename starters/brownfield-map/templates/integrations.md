# {{ .StageName }} - {{ .ProjectName }}

## Context
{{- range .Inputs }}
<artifact name="{{ .Name }}">
{{ .Content }}
</artifact>

{{- end }}

You will also receive repository files in `<file path="...">...</file>` blocks.

## Task
Produce `planning/codebase/INTEGRATIONS.md`.

## Output Requirements
List and describe:
- External APIs/services and what they’re used for
- Auth flows and credential handling (at a high level)
- Environment variables / config keys (if identifiable)
- Network boundaries (inbound/outbound)

## Output Format
Markdown with these exact headers:
- `# Integrations`
- `## External Services`
- `## Authentication & Secrets`
- `## Configuration Surface`
- `## Observability`
- `## Gaps / Uncertainties`

## Output Format Constraints
CRITICAL: Output ONLY the raw markdown content for `planning/codebase/INTEGRATIONS.md`.
Do not include conversational text or code fences.
