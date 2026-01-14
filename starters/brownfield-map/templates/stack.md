# {{ .StageName }} - {{ .ProjectName }}

You are mapping an existing (brownfield) codebase.

You will receive a bundle that contains repository files wrapped in tags like:

<file path="path/to/file.ext">
...content...
</file>

## Task
Produce `planning/codebase/STACK.md`.

## Output Requirements
Capture:
- Primary languages and runtime(s)
- Frameworks (web, backend, mobile, CLI)
- Package/dependency managers
- Datastores (DBs, caches, queues)
- Build and tooling (formatters, linters, generators)
- CI/CD signals (GitHub Actions, etc.)
- Local dev entrypoints (how to run, how to test)

## Output Format
Markdown with these exact headers:
- `# Stack`
- `## Runtimes & Languages`
- `## Frameworks & Major Libraries`
- `## Package & Build Tooling`
- `## Data Stores & Messaging`
- `## CI/CD & Automation`
- `## How To Run`
- `## How To Test`
- `## Gaps / Uncertainties`

## Rules
- Prefer facts from the files over assumptions.
- If something is unclear, state what evidence is missing.

## Output Format Constraints
CRITICAL: Output ONLY the raw markdown content for `planning/codebase/STACK.md`.
Do not include conversational text or code fences.
