package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"specfirst/internal/domain"
	"specfirst/internal/engine/templating"
	"specfirst/internal/repository"
)

type CompileOptions struct {
	Granularity    string
	MaxTasks       int
	PreferParallel bool
	RiskBias       string
}

func (app *Application) RequireStageDependencies(stage domain.Stage) error {
	for _, dep := range stage.DependsOn {
		if !app.State.IsStageCompleted(dep) {
			return fmt.Errorf("missing dependency: %s", dep)
		}
	}
	return nil
}

func (app *Application) CompilePrompt(stage domain.Stage, stageIDs []string, opts CompileOptions) (string, error) {
	var inputs []templating.Input
	if stage.Intent == "review" {
		artifacts, err := listAllArtifacts()
		if err != nil {
			return "", err
		}
		inputs = artifacts
	} else {
		inputs = make([]templating.Input, 0, len(stage.Inputs))
		for _, input := range stage.Inputs {
			// NOTE: we need dependent stage outputs from STATE to resolve artifacts accurately.
			// repository.ArtifactPathForInput needs the mapping.

			path, err := repository.ArtifactPathForInput(input, stage.DependsOn, stageIDs)
			if err != nil {
				return "", err
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return "", err
			}
			inputs = append(inputs, templating.Input{Name: input, Content: string(content)})
		}
	}

	// Apply options overrides
	if opts.Granularity != "" || opts.MaxTasks > 0 || opts.PreferParallel || opts.RiskBias != "" {
		if stage.Prompt == nil {
			stage.Prompt = &domain.PromptConfig{}
		}
		p := *stage.Prompt // shallow copy
		if opts.Granularity != "" {
			p.Granularity = opts.Granularity
		}
		if opts.MaxTasks > 0 {
			p.MaxTasks = opts.MaxTasks
		}
		if opts.PreferParallel {
			p.PreferParallel = opts.PreferParallel
		}
		if opts.RiskBias != "" {
			p.RiskBias = opts.RiskBias
		}
		stage.Prompt = &p
	}

	data := templating.Data{
		StageName:   stage.Name,
		ProjectName: app.Config.ProjectName,
		Inputs:      inputs,
		Outputs:     stage.Outputs,
		Intent:      stage.Intent,
		Language:    app.Config.Language,
		Framework:   app.Config.Framework,
		CustomVars:  app.Config.CustomVars,
		Constraints: app.Config.Constraints,

		StageType:      stage.Type,
		Prompt:         stage.Prompt,
		OutputContract: stage.Output,
		Epistemics:     app.State.Epistemics,
	}

	templatePath := repository.TemplatesPath(stage.Template)
	return templating.Render(templatePath, data)
}

func listAllArtifacts() ([]templating.Input, error) {
	artifactsRoot := repository.ArtifactsPath()
	info, err := os.Stat(artifactsRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return []templating.Input{}, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("artifacts path is not a directory: %s", artifactsRoot)
	}
	relPaths := []string{}
	err = filepath.WalkDir(artifactsRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(artifactsRoot, path)
		if err != nil {
			return err
		}
		relPaths = append(relPaths, rel)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(relPaths)

	inputs := make([]templating.Input, 0, len(relPaths))
	for _, rel := range relPaths {
		data, err := os.ReadFile(filepath.Join(artifactsRoot, rel))
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, templating.Input{Name: rel, Content: string(data)})
	}
	return inputs, nil
}
