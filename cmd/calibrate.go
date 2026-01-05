package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"

	"specfirst/internal/assets"
)

var calibrateMode string

var calibrateCmd = &cobra.Command{
	Use:   "calibrate <artifact>",
	Short: "Generate an epistemic annotation prompt for judgment calibration",
	Long: `Generate a prompt that helps calibrate confidence in a specification or artifact.

This command helps identify:
- What is known with high confidence
- What is assumed but not proven
- What is explicitly uncertain
- What would invalidate the spec

The output is a structured prompt suitable for AI assistants or human reviewers.

Modes:
  default     - Comprehensive calibration report (when no mode specified)
  confidence  - Classify statements into confidence tiers
  uncertainty - Surface ambiguity and underspecification
  unknowns    - Identify missing information`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		artifactPath := args[0]

		content, err := os.ReadFile(artifactPath)
		if err != nil {
			return fmt.Errorf("reading artifact %s: %w", artifactPath, err)
		}

		prompt, err := generateCalibratePrompt(string(content), artifactPath, calibrateMode)
		if err != nil {
			return err
		}

		prompt = applyMaxChars(prompt, stageMaxChars)
		formatted, err := formatPrompt(stageFormat, "calibrate", prompt)
		if err != nil {
			return err
		}

		if stageOut != "" {
			if err := writeOutput(stageOut, formatted); err != nil {
				return err
			}
		}
		_, err = cmd.OutOrStdout().Write([]byte(formatted))
		return err
	},
}

func init() {
	calibrateCmd.Flags().StringVarP(&calibrateMode, "mode", "m", "default", "calibration mode: default, confidence, uncertainty, unknowns")
}

// calibrateData holds template data for calibrate prompts
type calibrateData struct {
	ArtifactName string
	Content      string
}

func generateCalibratePrompt(content, path, mode string) (string, error) {
	var tmplContent string
	switch mode {
	case "default", "":
		tmplContent = assets.CalibrateDefaultTemplate
	case "confidence":
		tmplContent = assets.CalibrateConfidenceTemplate
	case "uncertainty":
		tmplContent = assets.CalibrateUncertaintyTemplate
	case "unknowns":
		tmplContent = assets.CalibrateUnknownsTemplate
	default:
		return "", fmt.Errorf("unknown mode: %s (valid: default, confidence, uncertainty, unknowns)", mode)
	}

	tmpl, err := template.New("calibrate").Parse(tmplContent)
	if err != nil {
		return "", fmt.Errorf("parsing calibrate template: %w", err)
	}

	data := calibrateData{
		ArtifactName: filepath.Base(path),
		Content:      content,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing calibrate template: %w", err)
	}

	return buf.String(), nil
}
