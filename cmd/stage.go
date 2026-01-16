package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

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

	if application.Config.Harness != "" && !stageDryRun {
		writer, ok := cmdOut.(io.Writer)
		if !ok {
			return fmt.Errorf("harness requires an io.Writer output")
		}
		return runHarness(application.Config.Harness, application.Config.HarnessArgs, formatted, writer)
	}

	if _, err := cmdOut.Write([]byte(formatted)); err != nil {
		return err
	}
	return nil
}

func runHarness(harness, args, prompt string, stdout io.Writer) error {
	argList, err := splitArgs(args)
	if err != nil {
		return err
	}
	cmd := exec.Command(harness, argList...)
	cmd.Stdin = strings.NewReader(prompt)
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	return cmd.Run()
}

func splitArgs(value string) ([]string, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	fields := []string{}
	current := strings.Builder{}
	inQuote := rune(0)
	escaped := false
	for _, r := range value {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false
		case r == '\\':
			escaped = true
		case inQuote != 0:
			if r == inQuote {
				inQuote = 0
			} else {
				current.WriteRune(r)
			}
		case r == '"' || r == '\'':
			inQuote = r
		case r == ' ' || r == '\t':
			if current.Len() > 0 {
				fields = append(fields, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if escaped {
		return nil, fmt.Errorf("invalid args: unfinished escape")
	}
	if inQuote != 0 {
		return nil, fmt.Errorf("invalid args: unterminated quote")
	}
	if current.Len() > 0 {
		fields = append(fields, current.String())
	}
	return fields, nil
}