# Code Review Starter

A staged, gated, and evidence-driven code review workflow designed to converge and eliminate infinite "hallucinated" issues.

## What This Demonstrates

- **Staged Convergence**: Architecture → File Review → Verification.
- **Gating**: Clear separation between structural critique and line-level nitpicks.
- **Anti-Hallucination**: A dedicated "Skeptic Pass" to filter out low-confidence findings.
- **Bounded Output**: Fixed budgets for findings to prevent endless review churn.

## Quick Start

1. **Initialize the starter**:
   ```bash
   specfirst init --starter code-review
   ```

2. **Define Scope**:
   ```bash
   # Analyze paths to review
   opencode run "$(specfirst scope ./src/core)" > scope.md
   specfirst complete scope ./scope.md
   ```

3. **Architecture Pass**:
   ```bash
   # Review structure (Architecture is frozen after this)
   opencode run "$(specfirst architecture-pass)" > architecture-findings.md
   specfirst complete architecture-pass ./architecture-findings.md
   ```

4. **File-Level Review**:
   ```bash
   # Focused review of individual files
   opencode run "$(specfirst file-review)" > file-findings.md
   specfirst complete file-review ./file-findings.md
   ```

5. **Skeptic Pass**:
   ```bash
   # Filter out hallucinations
   opencode run "$(specfirst skeptic-pass)" > verified-findings.md
   specfirst complete skeptic-pass ./verified-findings.md
   ```

6. **Final Report**:
   ```bash
   # Generate actionable report
   opencode run "$(specfirst report)" > review-report.md
   ```

## Workflow Details

### 1. Review Scope & Constraints
Sets the boundaries. You define exactly what files to look at and, crucially, a **budget** (e.g., "Max 2 findings per category").

### 2. Architecture Pass
A high-level view. No file-level nitpicks allowed. Once this stage is completed, the architecture is considered "frozen" for the rest of the review.

### 3. File-Level Review
Deep dive into the code. Every finding *must* cite a file and identifier. If it's a guess, it's labeled a "Question".

### 4. Skeptic Pass
The model is tasked with *disproving* its own findings. Only items with high confidence (≥ 70%) move forward.

### 5. Final Report
A prioritized, actionable summary of verified issues.

## When to Use This

- ✅ Large PRs that need structured evaluation.
- ✅ Onboarding to a new codebase.
- ✅ Auditing security-critical or performance-sensitive modules.
- ✅ When you find LLM reviews are "never-ending" or hallucinating issues.
- ❌ Small, obvious fixes.
- ❌ Rapid prototyping where deep review isn't yet needed.
