package app

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"specfirst/internal/domain"
	"specfirst/internal/engine/templating"
	"specfirst/internal/repository"
)

// ListTasks returns a validated list of tasks from the most recent decomposition.
func (app *Application) ListTasks() (domain.TaskList, []string, error) {
	// Find the decompose stage
	var decomposeStageID string
	for _, stage := range app.Protocol.Stages {
		if stage.Type == "decompose" {
			decomposeStageID = stage.ID
			break
		}
	}

	if decomposeStageID == "" {
		return domain.TaskList{}, nil, fmt.Errorf("no stage of type 'decompose' found in protocol %q", app.Protocol.Name)
	}

	if !app.State.IsStageCompleted(decomposeStageID) {
		return domain.TaskList{}, nil, fmt.Errorf("decompose stage %q has not been completed", decomposeStageID)
	}

	// Find the artifact for the decompose stage
	output, ok := app.State.StageOutputs[decomposeStageID]
	if !ok || len(output.Files) == 0 {
		return domain.TaskList{}, nil, fmt.Errorf("no artifacts found for decompose stage %q", decomposeStageID)
	}

	// Search through all artifacts to find a valid task list
	var taskList domain.TaskList
	var foundTaskList bool
	for _, file := range output.Files {
		artifactPath, err := repository.ArtifactAbsFromState(file)
		if err != nil {
			continue
		}

		content, err := os.ReadFile(artifactPath)
		if err != nil {
			continue
		}

		parsed, err := domain.ParseTaskList(string(content))
		if err == nil && len(parsed.Tasks) > 0 {
			taskList = parsed
			foundTaskList = true
			break
		}
	}

	if !foundTaskList {
		return domain.TaskList{}, nil, fmt.Errorf("no valid task list found in artifacts for decompose stage %q", decomposeStageID)
	}

	// Sort tasks by ID
	sort.Slice(taskList.Tasks, func(i, j int) bool {
		return taskList.Tasks[i].ID < taskList.Tasks[j].ID
	})

	warnings := taskList.Validate()
	return taskList, warnings, nil
}

// GenerateTaskPrompt generates an implementation prompt for a specific task.
func (app *Application) GenerateTaskPrompt(taskID string) (string, []string, error) {
	taskList, warnings, err := app.ListTasks()
	if err != nil {
		return "", nil, err
	}

	var targetTask *domain.Task
	for _, t := range taskList.Tasks {
		if t.ID == taskID {
			targetTask = &t
			break
		}
	}

	if targetTask == nil {
		return "", nil, fmt.Errorf("task %q not found in decomposition output", taskID)
	}

	// Find decompose stage again to get source for task prompt
	var decomposeStageID string
	for _, stage := range app.Protocol.Stages {
		if stage.Type == "decompose" {
			decomposeStageID = stage.ID
			break
		}
	}

	// Find the task_prompt stage that refers to this decompose stage
	var taskPromptStage domain.Stage
	var foundTaskPrompt bool
	for _, stg := range app.Protocol.Stages {
		if stg.Type == "task_prompt" && stg.Source == decomposeStageID {
			taskPromptStage = stg
			foundTaskPrompt = true
			break
		}
	}

	var artifactInputs []templating.Input
	if foundTaskPrompt {
		// Gather artifacts for the task_prompt stage (requirements, design, etc.)
		stageIDs := make([]string, 0, len(app.Protocol.Stages))
		for _, s := range app.Protocol.Stages {
			stageIDs = append(stageIDs, s.ID)
		}

		artifactInputs = make([]templating.Input, 0, len(taskPromptStage.Inputs))
		for _, input := range taskPromptStage.Inputs {
			path, err := repository.ArtifactPathForInput(input, taskPromptStage.DependsOn, stageIDs)
			if err != nil {
				continue
			}
			content, err := os.ReadFile(path)
			if err != nil {
				// Sev2 Fix: Don't silently ignore missing artifacts.
				// Log a warning so the user knows context is missing.
				fmt.Fprintf(os.Stderr, "Warning: failed to read artifact input %q: %v\n", input, err)
				continue
			}
			artifactInputs = append(artifactInputs, templating.Input{Name: input, Content: string(content)})
		}
	}

	promptText := generateTaskPrompt(*targetTask, artifactInputs, app.Config, app.Protocol)
	return promptText, warnings, nil
}

func generateTaskPrompt(t domain.Task, inputs []templating.Input, cfg domain.Config, proto domain.Protocol) string {
	var sb strings.Builder

	// Header
	sb.WriteString("---\n")
	sb.WriteString("intent: implementation\n")
	sb.WriteString("expected_output: code_diff\n")
	sb.WriteString("determinism: medium\n")
	sb.WriteString("allowed_creativity: low\n")
	sb.WriteString("---\n\n")

	sb.WriteString(fmt.Sprintf("# Implement %s: %s\n\n", t.ID, t.Title))

	if len(inputs) > 0 {
		sb.WriteString("## Context\n")
		for _, input := range inputs {
			sb.WriteString(fmt.Sprintf("<artifact name=\"%s\">\n", input.Name))
			sb.WriteString(input.Content)
			if !strings.HasSuffix(input.Content, "\n") {
				sb.WriteString("\n")
			}
			sb.WriteString("</artifact>\n\n")
		}
	}

	sb.WriteString("## Goal\n")
	sb.WriteString(t.Goal + "\n\n")

	if len(t.AcceptanceCriteria) > 0 {
		sb.WriteString("## Acceptance Criteria\n")
		for _, ac := range t.AcceptanceCriteria {
			sb.WriteString(fmt.Sprintf("- %s\n", ac))
		}
		sb.WriteString("\n")
	}

	if len(t.FilesTouched) > 0 {
		sb.WriteString("## Known Files\n")
		for _, f := range t.FilesTouched {
			sb.WriteString(fmt.Sprintf("- %s\n", f))
		}
		sb.WriteString("\n")
	}

	if len(t.Dependencies) > 0 {
		sb.WriteString("## Dependencies\n")
		sb.WriteString("This task depends on the completion of:\n")
		for _, dep := range t.Dependencies {
			sb.WriteString(fmt.Sprintf("- %s\n", dep))
		}
		sb.WriteString("\n")
	}

	if len(t.TestPlan) > 0 {
		sb.WriteString("## Test Plan\n")
		for _, tp := range t.TestPlan {
			sb.WriteString(fmt.Sprintf("- %s\n", tp))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Instructions\n")
	sb.WriteString("- Produce ONLY the minimal changes required.\n")
	sb.WriteString("- Maintain project standards and existing architecture.\n\n")

	sb.WriteString("## Expected Output\n")
	sb.WriteString("- Format: unified diff\n")
	sb.WriteString("- Scope: only listed files unless explicitly justified\n")
	sb.WriteString("- Tests: added or updated if behavior changes\n\n")

	sb.WriteString("## Assumptions\n")
	sb.WriteString("- (List explicitly before implementation if any)\n")

	return sb.String()
}
