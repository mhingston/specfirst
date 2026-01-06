package protocol

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRefactoringProtocolRepro(t *testing.T) {
	yamlContent := `name: "refactoring"
version: "1.0"

stages:
  - id: current-state
    name: Current State Analysis
    type: spec
    intent: exploration
    template: current-state.md
    outputs: [current-state.md]
    
  - id: goals
    name: Refactoring Goals
    type: spec
    intent: decision
    template: goals.md
    depends_on: [current-state]
    inputs: [current-state.md]
    outputs: [goals.md]
    
  - id: plan
    name: Refactoring Plan
    type: spec
    intent: planning
    template: plan.md
    depends_on: [current-state, goals]
    inputs: [current-state.md, goals.md]
    outputs: [plan.md]
    
  - id: execute
    name: Execute Refactoring
    type: spec
    intent: execution
    template: execute.md
    depends_on: [plan]
    inputs: [plan.md]
    outputs: []
    
  - id: verify
    name: Verification
    type: spec
    intent: review
    template: verify.md
    depends_on: [execute, plan]
    inputs: [plan.md]
    outputs: [verification-report.md]
`

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "repro.yaml")
	err := os.WriteFile(path, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Load(path)
	if err != nil {
		t.Fatalf("Failed to load protocol: %v", err)
	}
}
