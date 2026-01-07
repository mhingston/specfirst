package cmd

import (
	"specfirst/internal/app"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Run all non-blocking validations (lint, tasks, approvals, outputs)",
	RunE: func(cmd *cobra.Command, args []string) error {
		failOnWarnings, _ := cmd.Flags().GetBool("fail-on-warnings")

		application, err := app.Load(protocolFlag)
		if err != nil {
			return err
		}

		return application.Check(failOnWarnings)
	},
}

func init() {
	checkCmd.Flags().Bool("fail-on-warnings", false, "exit with code 1 if warnings are found")
}
