package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"specfirst/internal/app"
	"specfirst/internal/repository"
)

var trackCmd = &cobra.Command{
	Use:   "track",
	Short: "Manage parallel futures (tracks)",
}

var trackCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new track from current state",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		notes, _ := cmd.Flags().GetString("notes")

		application, err := app.Load(protocolFlag)
		if err != nil {
			return err
		}

		// Use repository for tracks
		mgr := repository.NewSnapshotRepository(repository.TracksPath())
		params := repository.CreateParams{
			Config:   application.Config,
			Protocol: application.Protocol,
			State:    application.State,
		}
		if err := mgr.Create(name, []string{"track"}, notes, params); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Created track %s\n", name)
		return nil
	},
}

var trackListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tracks",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := repository.NewSnapshotRepository(repository.TracksPath())
		tracks, err := mgr.List()
		if err != nil {
			return err
		}
		for _, t := range tracks {
			fmt.Fprintln(cmd.OutOrStdout(), t)
		}
		return nil
	},
}

var trackSwitchCmd = &cobra.Command{
	Use:   "switch <name>",
	Short: "Switch workspace to a specific track (restores it)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		force, _ := cmd.Flags().GetBool("force")

		if !force {
			if _, err := os.Stat(repository.ConfigPath()); err == nil {
				return fmt.Errorf("workspace has data; use --force to overwrite with track contents")
			}
		}

		mgr := repository.NewSnapshotRepository(repository.TracksPath())
		if err := mgr.Restore(name, force); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Switched to track %s\n", name)
		return nil
	},
}

var trackDiffCmd = &cobra.Command{
	Use:   "diff <track-a> <track-b>",
	Short: "Compare artifacts between two tracks",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := repository.NewSnapshotRepository(repository.TracksPath())
		added, removed, changed, err := mgr.Compare(args[0], args[1])
		if err != nil {
			return err
		}

		if len(added)+len(removed)+len(changed) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No differences detected.")
			return nil
		}
		if len(added) > 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "Added:")
			for _, item := range added {
				fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", item)
			}
		}
		if len(removed) > 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "Removed:")
			for _, item := range removed {
				fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", item)
			}
		}
		if len(changed) > 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "Changed:")
			for _, item := range changed {
				fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", item)
			}
		}
		return nil
	},
}

var trackMergeCmd = &cobra.Command{
	Use:   "merge <source-track>",
	Short: "Generate a merge plan to merge source track into current workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceTrack := args[0]

		application, err := app.Load(protocolFlag)
		if err != nil {
			return err
		}

		params := repository.CreateParams{
			Config:   application.Config,
			Protocol: application.Protocol,
			State:    application.State,
		}

		mgr := repository.NewSnapshotRepository(repository.TracksPath())

		currentSnapshot := "merge-target-temp"
		_ = os.RemoveAll(repository.TracksPath(currentSnapshot))

		if err := mgr.Create(currentSnapshot, []string{"temp"}, "Temporary snapshot for merge", params); err != nil {
			return fmt.Errorf("failed to snapshot current workspace for comparison: %w", err)
		}
		defer func() {
			_ = os.RemoveAll(repository.TracksPath(currentSnapshot))
		}()

		added, removed, changed, err := mgr.Compare(currentSnapshot, sourceTrack)
		if err != nil {
			return err
		}

		if len(added)+len(removed)+len(changed) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "Tracks are identical. Nothing to merge.")
			return nil
		}

		mergePromptPath := "MERGE_PLAN_PROMPT.md"
		promptContent := fmt.Sprintf(`# Merge Plan for %s into Current Workspace

## Context
- **Source Track**: %s

## Differences

### Added
%s

### Removed
%s

### Changed
%s

## Instructions
1. Review the differences above.
2. Generate a 'MERGE_PLAN.md' checklist to safely merge these changes.
3. Identify any conflicts or high-risk files.
`, sourceTrack, sourceTrack, formatList(added), formatList(removed), formatList(changed))

		if err := os.WriteFile(mergePromptPath, []byte(promptContent), 0644); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Merge prompt generated at %s\n", mergePromptPath)
		fmt.Fprintln(cmd.OutOrStdout(), "Run this prompt with your LLM to generate a granular merge strategy.")

		return nil
	},
}

func formatList(items []string) string {
	if len(items) == 0 {
		return "(none)"
	}
	res := ""
	for _, item := range items {
		res += fmt.Sprintf("- %s\n", item)
	}
	return res
}

func init() {
	rootCmd.AddCommand(trackCmd)
	trackCmd.AddCommand(trackCreateCmd)
	trackCmd.AddCommand(trackListCmd)
	trackCmd.AddCommand(trackSwitchCmd)
	trackCmd.AddCommand(trackDiffCmd)
	trackCmd.AddCommand(trackMergeCmd)

	trackCreateCmd.Flags().String("notes", "", "notes for the track")
	trackSwitchCmd.Flags().Bool("force", false, "force overwrite of existing workspace data")
}
