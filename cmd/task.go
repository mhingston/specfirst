package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"specfirst/internal/app"
	"specfirst/internal/engine/prompt"
)

var taskCmd = &cobra.Command{
	Use:   "task [task-id]",
	Short: "Generate implementation prompt for a specific task",
	Long: `Generate an implementation prompt for a specific task from a completed decomposition stage.
If no task ID is provided, it lists all available tasks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		application, err := app.Load(protocolFlag)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			// List tasks
			taskList, warnings, err := application.ListTasks()
			if err != nil {
				return err
			}

			fmt.Println("Available tasks:")
			for _, t := range taskList.Tasks {
				fmt.Printf("- %-10s: %s\n", t.ID, t.Title)
			}

			// Surface validation warnings
			if len(warnings) > 0 {
				fmt.Fprintln(os.Stderr, "\nWarnings:")
				for _, w := range warnings {
					fmt.Fprintf(os.Stderr, "- %s\n", w)
				}
			}
			return nil
		}

		taskID := args[0]
		promptText, warnings, err := application.GenerateTaskPrompt(taskID)
		if err != nil {
			// Surface validation warnings if any, then error
			if len(warnings) > 0 {
				fmt.Fprintln(os.Stderr, "Warnings:")
				for _, w := range warnings {
					fmt.Fprintf(os.Stderr, "- %s\n", w)
				}
				fmt.Fprintln(os.Stderr)
			}
			return err
		}

		// Surface validation warnings
		if len(warnings) > 0 {
			fmt.Fprintln(os.Stderr, "Warnings:")
			for _, w := range warnings {
				fmt.Fprintf(os.Stderr, "- %s\n", w)
			}
			fmt.Fprintln(os.Stderr)
		}

		// Surface ambiguity warnings in prompt
		if issues := prompt.ContainsAmbiguity(promptText); len(issues) > 0 {
			fmt.Fprintln(os.Stderr, "Ambiguity Warnings in generated prompt:")
			for _, issue := range issues {
				fmt.Fprintf(os.Stderr, "- %s\n", issue)
			}
			fmt.Fprintln(os.Stderr)
		}

		// Print Prompt
		fmt.Println(promptText)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(taskCmd)
}
