# AI CLI Integration

SpecFirst outputs prompts to stdout, making it composable with AI CLIs.

## Interactive vs. Non-Interactive Modes

Most AI CLIs support a **one-shot (non-interactive)** mode for automation and an **interactive** mode for refinement.

### 1. Interactive Refinement (Recommended)
To maintain an interactive session where you can refine the output, use your system clipboard or tools that support stdin-to-interactive:

```bash
# Copy prompt to clipboard and paste into your AI tool
specfirst requirements | pbcopy # macOS
specfirst requirements | xclip -sel clip # Linux

# Or use command substitution in the tool's interactive prompt
# (Works if the tool allows starting a session with an initial prompt)
copilot -p "$(specfirst requirements)" --allow-all-tools
```

### 2. One-Shot / Piped (Non-Interactive)
Use these for quick generations or scripting. Note that flags like `-p` or `--print` usually exit after one response.

### 2.1 Bundling Extra Context
If you need to include additional project files beyond the stage artifacts, use `specfirst bundle`.

```bash
# Bundle the prompt + selected files, then paste into your AI tool
specfirst bundle design --file "src/**" --exclude "src/**/generated/**" | pbcopy

# Tighter payload (no headings/report)
specfirst bundle design --file "src/**" --raw | pbcopy

# Shell-friendly heredoc you can eval/paste
specfirst bundle design --file "src/**" --shell

# Capture machine-readable report to stderr
specfirst bundle design --file "src/**" --raw --report-json - | cat >/tmp/bundle.md
```

```bash
# Claude Code (headless mode)
specfirst requirements | claude -p

# GitHub Copilot (non-interactive)
copilot -p "$(specfirst requirements)" --allow-all-tools

# Gemini CLI (non-interactive prompt)
opencode run "$(specfirst implementation)"

# Gemini CLI (via stdin)
specfirst implementation | opencode run
```

### 3. Pipelining Back to SpecFirst
You can pipe AI output directly into `specfirst complete` using `-` to read from stdin:

```bash
# Example: Generate requirements with Claude and complete the stage in one go
specfirst requirements | claude -p | specfirst complete requirements -
```

**For any tool that reads from files or requires non-interactive output for redirection**, use the tool's one-shot mode:

```bash
# Gemini one-shot to file
opencode run "$(specfirst requirements)" > requirements.md
```
