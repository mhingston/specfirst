package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"specfirst/internal/domain"

	"gopkg.in/yaml.v3"
)

var validStageIDPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// validStageTypes defines allowed stage types.
var validStageTypes = map[string]bool{
	"":            true, // Empty defaults to "spec"
	"spec":        true,
	"decompose":   true,
	"task_prompt": true,
}

// LoadProtocol reads and validates a protocol from a YAML file.
func LoadProtocol(path string) (domain.Protocol, error) {
	return LoadProtocolWithResolver(path, nil)
}

// LoadProtocolWithResolver loads a protocol with a custom resolver for imports.
func LoadProtocolWithResolver(path string, resolver func(name string) (domain.Protocol, error)) (domain.Protocol, error) {
	return loadProtocolWithDepth(path, resolver, make([]string, 0), make(map[string]domain.Protocol))
}

func loadProtocolWithDepth(path string, resolver func(name string) (domain.Protocol, error), visitedStack []string, processedCache map[string]domain.Protocol) (domain.Protocol, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return domain.Protocol{}, err
	}

	// 1. Cycle Detection
	for _, visited := range visitedStack {
		if visited == abs {
			return domain.Protocol{}, fmt.Errorf("circular protocol import detected: %s", abs)
		}
	}

	// 2. Memoization
	if cached, ok := processedCache[abs]; ok {
		return cached, nil
	}

	// 3. Load and Parse
	data, err := os.ReadFile(path)
	if err != nil {
		return domain.Protocol{}, err
	}

	var p domain.Protocol
	if err := yaml.Unmarshal(data, &p); err != nil {
		return domain.Protocol{}, err
	}

	// 4. Resolve Imports
	if len(p.Uses) > 0 {
		baseDir := filepath.Dir(path)
		newStack := append(visitedStack, abs)

		for i := len(p.Uses) - 1; i >= 0; i-- {
			importPath := p.Uses[i]
			var imported domain.Protocol
			var importErr error

			if resolver != nil {
				imported, importErr = resolver(importPath)
			} else {
				// Default: resolve relative to protocol directory
				importFile := filepath.Join(baseDir, importPath+".yaml")
				imported, importErr = loadProtocolWithDepth(importFile, nil, newStack, processedCache)
			}

			if importErr != nil {
				return domain.Protocol{}, fmt.Errorf("importing %q: %w", importPath, importErr)
			}

			// Prepend imported stages
			p.Stages = append(imported.Stages, p.Stages...)
			// Merge approvals
			p.Approvals = append(imported.Approvals, p.Approvals...)
		}
	}

	// Deduplicate Stages
	uniqueStages := make(map[string]bool)
	var dedupedStages []domain.Stage
	for i := len(p.Stages) - 1; i >= 0; i-- {
		s := p.Stages[i]
		if !uniqueStages[s.ID] {
			uniqueStages[s.ID] = true
			dedupedStages = append([]domain.Stage{s}, dedupedStages...)
		}
	}
	p.Stages = dedupedStages

	// Deduplicate Approvals
	uniqueApprovals := make(map[string]bool)
	var dedupedApprovals []domain.Approval
	for i := len(p.Approvals) - 1; i >= 0; i-- {
		a := p.Approvals[i]
		key := a.Stage + "::" + a.Role
		if !uniqueApprovals[key] {
			uniqueApprovals[key] = true
			dedupedApprovals = append([]domain.Approval{a}, dedupedApprovals...)
		}
	}
	p.Approvals = dedupedApprovals

	// Validation
	seen := make(map[string]bool)
	stageMap := make(map[string]domain.Stage)

	for _, stage := range p.Stages {
		if err := validateStageID(stage.ID); err != nil {
			return domain.Protocol{}, err
		}
		if err := validateTemplatePath(stage.Template); err != nil {
			return domain.Protocol{}, err
		}
		if err := validateStageType(stage.Type); err != nil {
			return domain.Protocol{}, fmt.Errorf("stage %q: %w", stage.ID, err)
		}
		if seen[stage.ID] {
			return domain.Protocol{}, fmt.Errorf("duplicate stage ID: %s", stage.ID)
		}
		seen[stage.ID] = true
		stageMap[stage.ID] = stage
	}

	// Dependency Validation
	for _, stage := range p.Stages {
		for _, dep := range stage.DependsOn {
			if dep == stage.ID {
				return domain.Protocol{}, fmt.Errorf("stage %q cannot depend on itself", stage.ID)
			}
			if !seen[dep] {
				return domain.Protocol{}, fmt.Errorf("stage %q depends on unknown stage %q", stage.ID, dep)
			}
		}
	}

	// Input Validation
	for _, stage := range p.Stages {
		for _, input := range stage.Inputs {
			if input == "" {
				continue
			}
			found := false

			for _, depID := range stage.DependsOn {
				depStage := stageMap[depID]
				for _, out := range depStage.Outputs {
					if matchPattern(out, input) {
						found = true
						break
					}
				}
				if found {
					break
				}
			}

			if !found && (strings.Contains(input, "/") || strings.Contains(input, string(os.PathSeparator))) {
				parts := strings.SplitN(filepath.ToSlash(input), "/", 2)
				stageID := parts[0]
				filename := parts[1]

				if targetStage, ok := stageMap[stageID]; ok {
					for _, out := range targetStage.Outputs {
						if matchPattern(out, filename) {
							found = true
							break
						}
					}
				}
			}

			if !found {
				return domain.Protocol{}, fmt.Errorf("stage %q: input %q not found in outputs of any dependency", stage.ID, input)
			}
		}
	}

	// Task Prompt Validation
	for _, stage := range p.Stages {
		if stage.Type == "task_prompt" && stage.Source != "" {
			if !seen[stage.Source] {
				return domain.Protocol{}, fmt.Errorf("stage %q references unknown source stage %q", stage.ID, stage.Source)
			}
			sourceStage, _ := p.StageByID(stage.Source)
			if sourceStage.Type != "decompose" {
				return domain.Protocol{}, fmt.Errorf("stage %q source must be a decompose stage", stage.ID)
			}
		}
	}

	// Cycle Detection
	for _, stage := range p.Stages {
		if err := checkCycles(p, stage.ID, []string{}); err != nil {
			return domain.Protocol{}, err
		}
	}

	// Approval Validation
	for _, approval := range p.Approvals {
		if strings.TrimSpace(approval.Stage) == "" {
			return domain.Protocol{}, fmt.Errorf("approval references empty stage")
		}
		if strings.TrimSpace(approval.Role) == "" {
			return domain.Protocol{}, fmt.Errorf("approval role is required for stage %q", approval.Stage)
		}
		if !seen[approval.Stage] {
			return domain.Protocol{}, fmt.Errorf("approval references unknown stage %q", approval.Stage)
		}
	}

	processedCache[abs] = p
	return p, nil
}

func checkCycles(p domain.Protocol, current string, path []string) error {
	for _, visited := range path {
		if visited == current {
			return fmt.Errorf("circular dependency detected: %v -> %s", path, current)
		}
	}
	path = append(path, current)

	stage, _ := p.StageByID(current)
	for _, dep := range stage.DependsOn {
		if err := checkCycles(p, dep, path); err != nil {
			return err
		}
	}
	return nil
}

func validateStageID(id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("stage id is required")
	}
	if id != strings.ToLower(id) {
		return fmt.Errorf("stage id %q must be lowercase", id)
	}
	if !validStageIDPattern.MatchString(id) {
		return fmt.Errorf("invalid stage id %q (must be alphanumeric and may contain hyphens or underscores)", id)
	}
	return nil
}

func matchPattern(pattern, file string) bool {
	if pattern == "" {
		return false
	}
	if strings.Contains(pattern, "*") {
		ok, _ := filepath.Match(pattern, file)
		return ok
	}
	return pattern == file
}

func validateTemplatePath(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("stage template is required")
	}
	clean := filepath.Clean(filepath.FromSlash(value))
	if filepath.IsAbs(clean) || clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("invalid template path: %s", value)
	}
	return nil
}

func validateStageType(stageType string) error {
	if !validStageTypes[stageType] {
		return fmt.Errorf("invalid stage type %q", stageType)
	}
	return nil
}
