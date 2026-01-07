package cmd

import (
	"specfirst/internal/app"
	"specfirst/internal/assets"
	"specfirst/internal/engine/prompt"
	"specfirst/internal/engine/templating"
	"specfirst/internal/repository"
)

type interactiveData struct {
	ProjectName  string
	ProtocolName string
	Stages       []stageSummary
	Language     string
	Framework    string
	CustomVars   map[string]string
	Constraints  map[string]string
}

type stageSummary struct {
	ID      string
	Name    string
	Intent  string
	Outputs []string
}

func runInteractive(cmdOut interface{ Write([]byte) (int, error) }) error {
	application, err := app.Load(protocolFlag)
	if err != nil {
		return err
	}

	stages := make([]stageSummary, 0, len(application.Protocol.Stages))
	for _, stage := range application.Protocol.Stages {
		stages = append(stages, stageSummary{
			ID:      stage.ID,
			Name:    stage.Name,
			Intent:  stage.Intent,
			Outputs: stage.Outputs,
		})
	}

	data := interactiveData{
		ProjectName:  application.Config.ProjectName,
		ProtocolName: application.Protocol.Name,
		Stages:       stages,
		Language:     application.Config.Language,
		Framework:    application.Config.Framework,
		CustomVars:   application.Config.CustomVars,
		Constraints:  application.Config.Constraints,
	}

	promptStr, err := templating.RenderInline(assets.InteractiveTemplate, data)
	if err != nil {
		return err
	}

	promptStr = prompt.ApplyMaxChars(promptStr, stageMaxChars)
	formatted, err := prompt.Format(stageFormat, "interactive", promptStr)
	if err != nil {
		return err
	}

	if stageOut != "" {
		if err := repository.WriteOutput(stageOut, formatted); err != nil {
			return err
		}
	}
	if _, err := cmdOut.Write([]byte(formatted)); err != nil {
		return err
	}
	return nil
}
