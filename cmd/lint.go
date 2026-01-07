package cmd

import (
	"github.com/spf13/cobra"

	"specfirst/internal/app"
)

var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Run non-blocking checks on the workspace",
	RunE: func(cmd *cobra.Command, args []string) error {
		application, err := app.Load(protocolFlag)
		if err != nil {
			return err
		}

		// Delegate to App.Check which handles all validation
		// failOnWarnings=false means return nil even if there are warnings
		return application.Check(false)
	},
}
