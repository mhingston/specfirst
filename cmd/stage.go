package cmd

import (
	"fmt"

	"specfirst/internal/app"
	"specfirst/internal/engine/prompt"
	"specfirst/internal/repository"
)

func runStage(cmdOut interface{ Write([]byte) (int, error) }, stageID string) error {
	application, err := app.Load(protocolFlag)
	if err != nil {
		return err
	}

	stage, ok := application.Protocol.StageByID(stageID)
	if !ok {
		return fmt.Errorf("unknown stage: %s", stageID)
	}

	if !stageNoStrict {
		if err := application.RequireStageDependencies(stage); err != nil {
			return err
		}
	}

	opts := app.CompileOptions{
		Granularity:    stageGranularity,
		MaxTasks:       stageMaxTasks,
		PreferParallel: stagePreferParallel,
		RiskBias:       stageRiskBias,
	}

	stageIDs := make([]string, 0, len(application.Protocol.Stages))
	for _, s := range application.Protocol.Stages {
		stageIDs = append(stageIDs, s.ID)
	}

	promptStr, err := application.CompilePrompt(stage, stageIDs, opts)
	if err != nil {
		return err
	}

	promptStr = prompt.ApplyMaxChars(promptStr, stageMaxChars)
	formatted, err := prompt.Format(stageFormat, stageID, promptStr)
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
