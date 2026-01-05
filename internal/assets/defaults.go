package assets

const DefaultProtocolName = "multi-stage"

const DefaultProtocolYAML = `name: "multi-stage"
version: "2.0"

stages:
  - id: requirements
    name: Requirements Gathering
    type: spec
    intent: exploration
    template: requirements.md
    depends_on: []
    outputs: [requirements.md]
    output:
      format: markdown
      sections:
        - Goals
        - Non-Goals
        - Assumptions
        - Constraints
        - Open Questions

  - id: design
    name: System Design
    type: spec
    intent: decision
    template: design.md
    depends_on: [requirements]
    inputs: [requirements.md]
    outputs: [design.md]
    prompt:
      intent: design_outline
      rules:
        - Do not propose implementation details
        - Prefer interfaces over concrete choices
    output:
      format: markdown
      sections:
        - Architecture
        - Components
        - Interfaces
        - Trade-offs

  - id: decompose
    name: Task Decomposition
    type: decompose
    intent: planning
    template: decompose.md
    depends_on: [requirements, design]
    inputs: [requirements.md, design.md]
    outputs: [tasks.yaml]
    prompt:
      intent: task_decomposition
      granularity: ticket
      max_tasks: 12
      prefer_parallel: true
      rules:
        - Prefer parallelizable tasks
        - Surface unknowns explicitly
        - Tasks must be independently reviewable
      required_fields:
        - id
        - title
        - goal
        - acceptance_criteria
        - dependencies
        - files_touched
        - risk_level
        - estimated_scope
        - test_plan
    output:
      format: yaml

  - id: implement
    name: Implementation
    type: task_prompt
    intent: execution
    template: implementation.md
    depends_on: [requirements, design, decompose]
    inputs: [requirements.md, design.md, tasks.yaml]
    source: decompose
    outputs: []
`

const RequirementsTemplate = `# {{ .StageName }} - {{ .ProjectName }}

## Context
{{- if .Inputs }}
{{- range .Inputs }}
<artifact name="{{ .Name }}">
{{ .Content }}
</artifact>

{{- end }}
{{- else }}
(No prior artifacts)
{{- end }}

## Task
Gather requirements, ask clarifying questions, and enumerate constraints.

{{- if .Outputs }}
## Output Requirements
{{- range .Outputs }}
- {{ . }}
{{- end }}
{{- end }}
`

const DesignTemplate = `# {{ .StageName }} - {{ .ProjectName }}

## Context
{{- if .Inputs }}
{{- range .Inputs }}
<artifact name="{{ .Name }}">
{{ .Content }}
</artifact>

{{- end }}
{{- end }}

## Task
Produce a system design that satisfies the requirements. Make binding decisions and note trade-offs.

{{- if .Outputs }}
## Output Requirements
{{- range .Outputs }}
- {{ . }}
{{- end }}
{{- end }}
`

const ImplementationTemplate = `# {{ .StageName }} - {{ .ProjectName }}

## Context
{{- if .Inputs }}
{{- range .Inputs }}
<artifact name="{{ .Name }}">
{{ .Content }}
</artifact>

{{- end }}
{{- end }}

## Task
Implement the specified system. Produce concrete artifacts only.

{{- if .Outputs }}
## Output Requirements
{{- range .Outputs }}
- {{ . }}
{{- end }}
{{- end }}
`

const DecomposeTemplate = `# {{ .StageName }} - {{ .ProjectName }}

## Context
{{- if .Inputs }}
{{- range .Inputs }}
<artifact name="{{ .Name }}">
{{ .Content }}
</artifact>

{{- end }}
{{- end }}

## Task
You are decomposing an approved design into implementation tasks.

### Rules
- Tasks must be independently reviewable.
- If assumptions are required, state them explicitly.
- If any acceptance criteria cannot be verified, add a task to define verification.

## Output Requirements
- Format: YAML
- Each task must include: id, title, goal, acceptance_criteria, dependencies, files_touched, risk_level, estimated_scope, test_plan
`

const DefaultConfigTemplate = `project_name: %s
protocol: %s
language: ""
framework: ""


custom_vars: {}
constraints: {}
`

const InteractiveTemplate = `# Interactive Spec Session - {{ .ProjectName }}

You are facilitating a spec-first workflow. Ask any clarifying questions first,
then produce the requested artifacts for each stage in order.

## Workflow
Protocol: {{ .ProtocolName }}

{{- range .Stages }}
- {{ .ID }}: {{ .Name }} (intent: {{ .Intent }})
  outputs:
  {{- if .Outputs }}
  {{- range .Outputs }}
    - {{ . }}
  {{- end }}
  {{- else }}
    - (none)
  {{- end }}
{{- end }}

## Instructions
- Start with any questions needed to clarify requirements or constraints.
- Then generate artifacts for each stage, in order, using the filenames listed.
- Do not invent new stages or outputs.
- Keep outputs structured and concise.

## Project Context
- Language: {{ .Language }}
- Framework: {{ .Framework }}
{{- if .Constraints }}
- Constraints:
{{- range $key, $value := .Constraints }}
  - {{ $key }}: {{ $value }}
{{- end }}
{{- end }}
{{- if .CustomVars }}
- Custom Vars:
{{- range $key, $value := .CustomVars }}
  - {{ $key }}: {{ $value }}
{{- end }}
{{- end }}
`

// Calibrate templates for epistemic annotation

const CalibrateDefaultTemplate = `# Epistemic Calibration

You are performing *judgment calibration* on the artifact(s) below.
Do not propose implementation. Do not rewrite the artifact. Do not add requirements.
Your job is to annotate epistemic status so a human can decide what to do next.

## Inputs

<artifact name="{{ .ArtifactName }}">
{{ .Content }}
</artifact>

## Task

Produce a calibration report with the following sections **exactly**:

### 1) High-Confidence Claims
List statements that are clearly supported by the artifact.
- Each bullet must be a *specific claim*.
- Include the supporting evidence by pointing to the relevant part (quote a short phrase, <= 20 words).

### 2) Assumptions (Unproven but Required)
List assumptions the artifact depends on.
For each:
- Assumption:
- Why it matters:
- What breaks if false:
- How to validate (simple check, experiment, or question):

### 3) Uncertainties (Ambiguous or Underspecified)
List ambiguities, underspecified areas, or multiple valid interpretations.
For each:
- Uncertainty:
- Competing interpretations (if any):
- Minimum clarification question:

### 4) Unknowns / Missing Information
List information that is missing but likely necessary to proceed safely.
For each:
- Missing info:
- Impact if ignored:
- Best way to obtain it:

### 5) Red Flags
Identify risks of proceeding without further clarification.
Examples: inconsistent constraints, contradictions, hidden coupling, unverifiable requirements, unsafe defaults.
For each:
- Red flag:
- Severity: low | medium | high
- Suggested mitigation:

### 6) Decision Checklist (for a human)
Create a short checklist (5–12 items) of decisions a human must explicitly make before proceeding.

## Output Rules
- Be concise and concrete.
- Do not invent details not present in the artifact.
- If the artifact is too vague to evaluate, say so in section 3 and 4 and explain why.
`

const CalibrateConfidenceTemplate = `# Calibration Mode: Confidence

You are to identify what can be treated as reliable vs tentative in the artifact(s) below.
Do not add requirements. Do not propose a solution. Do not write code.

## Inputs

<artifact name="{{ .ArtifactName }}">
{{ .Content }}
</artifact>

## Task

Classify statements into confidence tiers. Use the format below **exactly**.

### A) High Confidence
Only include claims that are explicitly stated and internally consistent.
For each bullet:
- Claim:
- Evidence (<= 20 word quote or pinpoint reference):

### B) Medium Confidence
Claims that are implied or plausible but not fully pinned down.
For each bullet:
- Claim:
- Why not high confidence:
- What evidence would upgrade it:

### C) Low Confidence
Claims that are speculative, ambiguous, or require external facts.
For each bullet:
- Claim:
- What's missing:
- What could falsify it quickly:

### D) Confidence Killers
List the top 3–7 issues that reduce confidence in the artifact overall.
For each:
- Issue:
- Where it appears:
- Fastest fix:

## Output Rules
- Prefer fewer, stronger bullets over many weak ones.
- Do not rely on external knowledge; treat the artifact as the source of truth.
`

const CalibrateUncertaintyTemplate = `# Calibration Mode: Uncertainty

Your job is to surface ambiguity and underspecification that could cause misimplementation.
Do not propose designs. Do not generate code. Do not rewrite the artifact.

## Inputs

<artifact name="{{ .ArtifactName }}">
{{ .Content }}
</artifact>

## Task

Generate an "Uncertainty Register" using the table-like format below.
Include **only** items that materially affect correct implementation.

For each item, provide:

- ID: U1, U2, ...
- Unclear statement or area:
- Why it's ambiguous (what could it mean?):
- Two plausible interpretations:
- Consequence if chosen wrong:
- Minimal clarifying question (one sentence):
- Suggested default (only if the artifact already implies one; otherwise say "none"):

Finally, add:

### Ambiguity Hotspots
List the top 3 sections most likely to be misread and why.

## Output Rules
- Do not invent interpretations that are wildly unrelated; keep them plausible.
- Keep clarifying questions minimal and decision-forcing.
`

const CalibrateUnknownsTemplate = `# Calibration Mode: Unknowns

Your job is to identify missing information required to proceed safely.
Do not propose implementation. Do not assume external context unless it is explicitly provided.

## Inputs

<artifact name="{{ .ArtifactName }}">
{{ .Content }}
</artifact>

## Task

Produce a "Missing Information Inventory" with the following sections:

### 1) Missing Decisions
List decisions that must be made (by a human) but are not made in the artifact.
For each:
- Decision:
- Options (if implied):
- Who should decide:
- Latest responsible moment:

### 2) Missing Constraints
List constraints that are necessary for correctness/safety but absent.
Examples: performance budgets, data retention, auth model, platform limits, compatibility.
For each:
- Constraint:
- Why it matters:
- How to elicit it (question to ask):

### 3) Missing Operational Reality
List operational details needed for a workable system.
Examples: deployment target, observability expectations, failure handling, rollback strategy.
For each:
- Missing operational detail:
- What breaks without it:
- Minimal question:

### 4) Missing Acceptance Criteria
List what would make this "done" in a verifiable way.
For each:
- Acceptance criterion:
- How to test or observe it:

### 5) Minimal Next-Step Questions
Provide 5–15 short questions that, if answered, would reduce unknowns the most.
Order by impact.

## Output Rules
- Do not fill in missing items; name them.
- Prefer questions that force a concrete choice.
`
