package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"specfirst/internal/assets"
	"specfirst/internal/starter"
	"specfirst/internal/store"
)

var (
	initStarter string
	initChoose  bool
	initForce   bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a SpecFirst workspace",
	Long: `Initialize a SpecFirst workspace in the current directory.

Creates .specfirst/ with default protocol, templates, and config.

Options:
  --starter <name>  Initialize with a specific starter kit workflow
  --choose          Interactively select a starter kit
  --force           Overwrite existing templates/protocols (only with --starter or --choose)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create workspace directories
		if err := ensureDir(store.SpecPath()); err != nil {
			return err
		}
		if err := ensureDir(store.ArtifactsPath()); err != nil {
			return err
		}
		if err := ensureDir(store.GeneratedPath()); err != nil {
			return err
		}
		if err := ensureDir(store.ProtocolsPath()); err != nil {
			return err
		}
		if err := ensureDir(store.TemplatesPath()); err != nil {
			return err
		}
		if err := ensureDir(store.ArchivesPath()); err != nil {
			return err
		}

		// Handle starter selection
		selectedStarter := initStarter
		if initChoose {
			chosen, err := interactiveSelectStarter(cmd)
			if err != nil {
				return err
			}
			selectedStarter = chosen
		}

		// If a starter is selected, apply it after base setup
		if selectedStarter != "" {
			// 1. Create basic config first if missing (so Apply can merge defaults into it)
			configPath := store.ConfigPath()
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				projectName := filepath.Base(mustGetwd())
				cfg := fmt.Sprintf(assets.DefaultConfigTemplate, projectName, "temp") // temporary protocol name, will be updated by Apply
				if err := os.WriteFile(configPath, []byte(cfg), 0644); err != nil {
					return err
				}
			}

			// 2. Apply the starter (copies protocol, templates, skills AND updates config with defaults)
			if err := starter.Apply(selectedStarter, initForce, true); err != nil {
				return fmt.Errorf("applying starter %q: %w", selectedStarter, err)
			}
		} else {
			// Default behavior: write default protocol and templates
			if err := writeIfMissing(store.ProtocolsPath(assets.DefaultProtocolName+".yaml"), assets.DefaultProtocolYAML); err != nil {
				return err
			}
			if err := writeIfMissing(store.TemplatesPath("requirements.md"), assets.RequirementsTemplate); err != nil {
				return err
			}
			if err := writeIfMissing(store.TemplatesPath("design.md"), assets.DesignTemplate); err != nil {
				return err
			}
			if err := writeIfMissing(store.TemplatesPath("implementation.md"), assets.ImplementationTemplate); err != nil {
				return err
			}
			if err := writeIfMissing(store.TemplatesPath("decompose.md"), assets.DecomposeTemplate); err != nil {
				return err
			}

			configPath := store.ConfigPath()
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				projectName := filepath.Base(mustGetwd())
				protoName := assets.DefaultProtocolName
				if protocolFlag != "" {
					protoName = protocolFlag
				}
				cfg := fmt.Sprintf(assets.DefaultConfigTemplate, projectName, protoName)
				if err := os.WriteFile(configPath, []byte(cfg), 0644); err != nil {
					return err
				}
			}
		}

		// Write state file if missing
		statePath := store.StatePath()
		if _, err := os.Stat(statePath); os.IsNotExist(err) {
			if err := os.WriteFile(statePath, []byte("{}\n"), 0644); err != nil {
				return err
			}
		}

		if selectedStarter != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Initialized .specfirst workspace with starter %q\n", selectedStarter)
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "Initialized .specfirst workspace")
		}
		return nil
	},
}

// interactiveSelectStarter prompts the user to select a starter.
func interactiveSelectStarter(cmd *cobra.Command) (string, error) {
	starters, err := starter.List()
	if err != nil {
		return "", err
	}

	if len(starters) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No starters found. Using default protocol.")
		return "", nil
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Available starters:")
	fmt.Fprintln(cmd.OutOrStdout(), "  0) [default] Use default multi-stage protocol")
	for i, s := range starters {
		fmt.Fprintf(cmd.OutOrStdout(), "  %d) %s\n", i+1, s.Name)
	}

	fmt.Fprint(cmd.OutOrStdout(), "\nSelect a starter [0]: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" || input == "0" {
		return "", nil // Use default
	}

	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(starters) {
		return "", fmt.Errorf("invalid selection: %s", input)
	}

	return starters[choice-1].Name, nil
}

func writeIfMissing(path string, content string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	return os.WriteFile(path, []byte(content), 0644)
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "project"
	}
	return wd
}

func init() {
	initCmd.Flags().StringVar(&initStarter, "starter", "", "initialize with a specific starter kit")
	initCmd.Flags().BoolVar(&initChoose, "choose", false, "interactively select a starter kit")
	initCmd.Flags().BoolVar(&initForce, "force", false, "overwrite existing templates/protocols")
}
