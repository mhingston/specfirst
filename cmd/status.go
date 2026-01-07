package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"specfirst/internal/app"
	"specfirst/internal/repository"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show workflow status",
	RunE: func(cmd *cobra.Command, args []string) error {
		application, err := app.Load(protocolFlag)
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Project: %s\n", application.Config.ProjectName)
		fmt.Fprintf(cmd.OutOrStdout(), "Protocol: %s\n", application.Protocol.Name)
		fmt.Fprintf(cmd.OutOrStdout(), "State: %s\n", repository.StatePath())

		if len(application.State.CompletedStages) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "Completed stages: (none)")
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "Completed stages: %v\n", application.State.CompletedStages)
		}

		if application.State.CurrentStage != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Current stage: %s\n", application.State.CurrentStage)
		}
		return nil
	},
}
