package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"specfirst/internal/app"
	"specfirst/internal/bundle"
	"specfirst/internal/engine/prompt"
	"specfirst/internal/repository"
)

var (
	bundleFiles      []string
	bundleExcludes   []string
	bundleMaxFiles   int
	bundleMaxBytes   int64
	bundleMaxPerFile int64
	bundleNoDefaults bool
	bundleNoReport   bool
	bundleRaw        bool
	bundleShell      bool
	bundleReportJSON string
)

func renderBundleBody(stageID string, promptStr string, files []bundle.File, raw bool) string {
	var b strings.Builder
	if raw {
		fmt.Fprintf(&b, "<prompt stage=\"%s\">\n%s\n</prompt>\n\n", stageID, promptStr)
		for _, f := range files {
			fmt.Fprintf(&b, "<file path=\"%s\">\n%s\n</file>\n\n", f.Path, f.Content)
		}
		return b.String()
	}

	fmt.Fprintf(&b, "## Prompt\n\n")
	fmt.Fprintf(&b, "<prompt stage=\"%s\">\n%s\n</prompt>\n\n", stageID, promptStr)

	fmt.Fprintf(&b, "## Files\n\n")
	for _, f := range files {
		fmt.Fprintf(&b, "<file path=\"%s\">\n%s\n</file>\n\n", f.Path, f.Content)
	}

	return b.String()
}

func escapeSingleQuotes(value string) string {
	return strings.ReplaceAll(value, "'", "'\"'\"'")
}

func heredocDelimiter(body string) string {
	base := "SPECFIRST_BUNDLE_EOF"
	delimiter := base
	for i := 0; ; i++ {
		if i > 0 {
			delimiter = fmt.Sprintf("%s_%d", base, i)
		}
		line := delimiter + "\n"
		if !strings.Contains(body, "\n"+delimiter+"\n") &&
			!strings.HasPrefix(body, line) &&
			!strings.HasSuffix(body, "\n"+delimiter) &&
			body != delimiter {
			return delimiter
		}
	}
}

type bundleReportJSONPayload struct {
	Stage         string `json:"stage"`
	Protocol      string `json:"protocol"`
	PromptChars   int    `json:"prompt_chars"`
	IncludedFiles int    `json:"included_files"`
	IncludedBytes int64  `json:"included_bytes"`

	Limits struct {
		MaxFiles     int   `json:"max_files"`
		MaxBytes     int64 `json:"max_bytes"`
		MaxFileBytes int64 `json:"max_file_bytes"`
	} `json:"limits"`

	Skipped struct {
		Excluded  int `json:"excluded"`
		TooLarge  int `json:"too_large"`
		OverLimit int `json:"over_limit"`
	} `json:"skipped"`

	Files []struct {
		Path  string `json:"path"`
		Bytes int64  `json:"bytes"`
	} `json:"files"`

	MissingFiles []string `json:"missing_files,omitempty"`
}

func buildBundleReportJSON(stageID string, protocolName string, promptStr string, files []bundle.File, report bundle.Report) ([]byte, error) {
	payload := bundleReportJSONPayload{
		Stage:         stageID,
		Protocol:      protocolName,
		PromptChars:   len([]rune(promptStr)),
		IncludedFiles: report.IncludedFiles,
		IncludedBytes: report.IncludedBytes,
		MissingFiles:  report.MissingLiterals,
	}
	payload.Limits.MaxFiles = bundleMaxFiles
	payload.Limits.MaxBytes = bundleMaxBytes
	payload.Limits.MaxFileBytes = bundleMaxPerFile
	payload.Skipped.Excluded = report.SkippedByExclude
	payload.Skipped.TooLarge = report.SkippedTooLarge
	payload.Skipped.OverLimit = report.SkippedOverLimit

	payload.Files = make([]struct {
		Path  string `json:"path"`
		Bytes int64  `json:"bytes"`
	}, 0, len(files))
	for _, f := range files {
		payload.Files = append(payload.Files, struct {
			Path  string `json:"path"`
			Bytes int64  `json:"bytes"`
		}{Path: f.Path, Bytes: f.Bytes})
	}

	return json.MarshalIndent(payload, "", "  ")
}

var bundleCmd = &cobra.Command{
	Use:   "bundle <stage-id>",
	Short: "Bundle a stage prompt with extra files for pasting into an LLM",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return stageIDCompletions(cmd, args, toComplete)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		stageID := args[0]
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

		files, report, err := bundle.Collect(bundle.Options{
			IncludePatterns: bundleFiles,
			ExcludePatterns: bundleExcludes,
			MaxFiles:        bundleMaxFiles,
			MaxTotalBytes:   bundleMaxBytes,
			MaxFileBytes:    bundleMaxPerFile,
			DefaultExcludes: !bundleNoDefaults,
		})
		if err != nil {
			if errors.Is(err, bundle.ErrNoFilesSelected) {
				return fmt.Errorf("no files matched; adjust --file/--exclude or raise limits")
			}
			return err
		}

		if bundleReportJSON != "" {
			data, err := buildBundleReportJSON(stageID, application.Protocol.Name, promptStr, files, report)
			if err != nil {
				return err
			}
			if bundleReportJSON == "-" {
				fmt.Fprintln(cmd.ErrOrStderr(), string(data))
			} else {
				if err := repository.WriteOutput(bundleReportJSON, string(data)+"\n"); err != nil {
					return err
				}
			}
		}

		bundleBody := renderBundleBody(stageID, promptStr, files, bundleRaw)
		if bundleShell {
			delimiter := heredocDelimiter(bundleBody)
			fmt.Fprintf(cmd.OutOrStdout(), "SPECFIRST_STAGE='%s'\nSPECFIRST_BUNDLE=$(cat <<'%s'\n%s\n%s\n)\n", escapeSingleQuotes(stageID), delimiter, bundleBody, delimiter)
			return nil
		}

		out := cmd.OutOrStdout()
		if !bundleNoReport && !bundleRaw {
			fmt.Fprintf(out, "# SpecFirst Bundle\n\n")
			fmt.Fprintf(out, "- stage: `%s`\n", stageID)
			fmt.Fprintf(out, "- protocol: `%s`\n", application.Protocol.Name)
			fmt.Fprintf(out, "- prompt_chars: %d\n", len([]rune(promptStr)))
			fmt.Fprintf(out, "- files: %d (max %d)\n", report.IncludedFiles, bundleMaxFiles)
			fmt.Fprintf(out, "- file_bytes: %d (max %d)\n", report.IncludedBytes, bundleMaxBytes)
			fmt.Fprintf(out, "- max_file_bytes: %d\n", bundleMaxPerFile)
			fmt.Fprintf(out, "- skipped: excluded=%d too_large=%d over_limit=%d\n\n", report.SkippedByExclude, report.SkippedTooLarge, report.SkippedOverLimit)

			fmt.Fprintf(out, "## Included Files\n")
			for _, f := range files {
				fmt.Fprintf(out, "- `%s` (%d bytes)\n", f.Path, f.Bytes)
			}
			fmt.Fprintln(out)

			if len(report.MissingLiterals) > 0 {
				sort.Strings(report.MissingLiterals)
				fmt.Fprintf(out, "## Missing Files\n")
				for _, p := range report.MissingLiterals {
					fmt.Fprintf(out, "- `%s`\n", p)
				}
				fmt.Fprintln(out)
			}
		}

		fmt.Fprint(out, bundleBody)
		return nil
	},
}

func init() {
	bundleCmd.Flags().StringArrayVar(&bundleFiles, "file", nil, "include files by glob (supports **), relative to project root")
	bundleCmd.Flags().StringArrayVar(&bundleExcludes, "exclude", nil, "exclude files by glob (supports **), relative to project root")
	bundleCmd.Flags().IntVar(&bundleMaxFiles, "max-files", 50, "maximum files to include")
	bundleCmd.Flags().Int64Var(&bundleMaxBytes, "max-bytes", 250_000, "maximum total bytes to include")
	bundleCmd.Flags().Int64Var(&bundleMaxPerFile, "max-file-bytes", 100_000, "maximum bytes per file")
	bundleCmd.Flags().BoolVar(&bundleNoDefaults, "no-default-excludes", false, "disable default excludes (.git, .specfirst, etc.)")
	bundleCmd.Flags().BoolVar(&bundleNoReport, "no-report", false, "omit bundle summary report")
	bundleCmd.Flags().BoolVar(&bundleRaw, "raw", false, "emit only <prompt>/<file> blocks (no headings/report)")
	bundleCmd.Flags().BoolVar(&bundleShell, "shell", false, "emit a bash heredoc assignment to SPECFIRST_BUNDLE")
	bundleCmd.Flags().StringVar(&bundleReportJSON, "report-json", "", "write a machine-readable JSON report to a file (or '-' for stderr)")

	_ = bundleCmd.MarkFlagRequired("file")

	_ = bundleCmd.RegisterFlagCompletionFunc("file", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Suggest a few common include patterns; shell can still complete paths.
		candidates := []string{"src/**", "internal/**", "cmd/**", "**/*.go", "**/*.ts", "**/*.tsx", "**/*.py", "README*", ".github/workflows/**"}
		return filterPrefix(candidates, toComplete), cobra.ShellCompDirectiveDefault
	})
	_ = bundleCmd.RegisterFlagCompletionFunc("exclude", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		candidates := []string{".git/**", ".specfirst/**", "node_modules/**", "dist/**", "tmp/**", "**/*.min.*", "**/*.lock"}
		return filterPrefix(candidates, toComplete), cobra.ShellCompDirectiveDefault
	})
	_ = bundleCmd.RegisterFlagCompletionFunc("report-json", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return filterPrefix([]string{"-"}, toComplete), cobra.ShellCompDirectiveDefault
	})
}
