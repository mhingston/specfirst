# Brownfield Map Starter

Generate durable, reusable documentation for an existing codebase.

This starter is inspired by the “map-codebase” workflow in GSD-style systems: create a small set of stable reference docs once, then reuse them as context for future feature/bugfix/refactor prompts.

## Outputs

This workflow generates artifacts under `planning/codebase/`:

- `STACK.md` — languages, frameworks, build/run/test commands
- `STRUCTURE.md` — directory map and where to change things
- `ARCHITECTURE.md` — components and data flow
- `CONVENTIONS.md` — naming, layering, error/logging patterns
- `TESTING.md` — how tests are organized and run
- `INTEGRATIONS.md` — external services and config surface
- `CONCERNS.md` — risks, tech debt hotspots, suggested next checks
- `CODEBASE.md` — single, reusable summary document
- `STATE.md` — living memory / current focus notes

## Recommended Usage (with `specfirst bundle`)

1. Install the starter:

```bash
specfirst init --starter brownfield-map
```

2. For each stage, bundle the prompt plus representative code files, run your AI tool, then complete the stage.

Example (OpenCode one-shot):

```bash
# 1) Stack survey
specfirst bundle map-stack \
  --file "README*" \
  --file "go.mod" \
  --file "package.json" \
  --file "pyproject.toml" \
  --file "Cargo.toml" \
  --file ".github/workflows/**" \
  --file "cmd/**" \
  --file "internal/**" \
  --file "src/**" \
  --exclude "**/*.min.*" \
  --exclude "**/*.lock" \
  > /tmp/map-stack.prompt.md

opencode run "$(cat /tmp/map-stack.prompt.md)" > STACK.md
specfirst complete map-stack STACK.md
```

3. Repeat for the other mapping stages:

```bash
specfirst bundle map-structure --file "**" --exclude "node_modules/**" > /tmp/map-structure.prompt.md
opencode run "$(cat /tmp/map-structure.prompt.md)" > STRUCTURE.md
specfirst complete map-structure STRUCTURE.md

# …and so on for:
# map-architecture, map-conventions, map-testing, map-integrations, map-concerns, map-summary, map-state
```

## Tips

- Start with a focused file set; expand only when the model needs more evidence.
- Prefer `--exclude` to keep bundles small and relevant.
- Use `--raw` if you want the tightest possible payload for your AI tool.
