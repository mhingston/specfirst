# {{ .StageName }} - {{ .ProjectName }}

## Context
{{- range .Inputs }}
<artifact name="{{ .Name }}">
{{ .Content }}
</artifact>

{{- end }}

You may also receive repository files in `<file path="...">...</file>` blocks.

## Task
Produce `planning/codebase/CODEBASE.md` as a single, reusable “memory” document.

## Output Requirements
Your output MUST:
- Summarize the system in a way that helps future implementation work.
- Be concise but information-dense.
- Include concrete "where to change things" guidance.
- Capture unknowns and what to check next.

## Output Format
Markdown with these exact headers:
- `# Codebase Summary`
- `## One-Line Description`
- `## Architecture (Short)`
- `## Key Directories`
- `## How To Run`
- `## How To Test`
- `## Integrations`
- `## Conventions To Follow`
- `## Risks / Fragile Areas`
- `## Open Questions`
- `## Next Suggested Checks`

## Rules
- Use evidence from the mapping docs; don’t invent.
- When stating a convention or risk, reference the relevant doc name (e.g., “see CONVENTIONS.md”).

## Output Format Constraints
CRITICAL: Output ONLY the raw markdown content for `planning/codebase/CODEBASE.md`.
Do not include conversational text or code fences.
