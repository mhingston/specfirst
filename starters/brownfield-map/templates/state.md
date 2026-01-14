# {{ .StageName }} - {{ .ProjectName }}

## Context
You have a stable codebase reference doc set (STACK/STRUCTURE/ARCHITECTURE/… plus CODEBASE.md).

{{- range .Inputs }}
<artifact name="{{ .Name }}">
{{ .Content }}
</artifact>

{{- end }}

## Task
Produce `planning/codebase/STATE.md` as a **living memory** document for future work.

This is NOT a restatement of CODEBASE.md. It is a change-oriented working note: what matters right now, what’s risky, what’s unknown, and what to check before changing things.

## Output Requirements
Include:
- What we believe is true about the system today
- Current priorities / current goals (even if tentative)
- Known risks and fragile areas to treat carefully
- Open questions and how to answer them (which files/commands to inspect)
- Suggested next actions for a typical “add feature / fix bug” task

## Output Format
Markdown with these exact headers:
- `# State`
- `## Current Snapshot`
- `## Current Goals`
- `## Risks & Fragile Areas`
- `## Open Questions`
- `## Next Actions`

## Rules
- Keep it short (aim for ~1–2 pages).
- Prefer actionable bullets over prose.

## Output Format Constraints
CRITICAL: Output ONLY the raw markdown content for `planning/codebase/STATE.md`.
Do not include conversational text or code fences.
