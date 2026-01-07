// Package starter provides discovery and application of starter kit workflows.
package starter

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"

	"specfirst/internal/repository"
)

// EmbeddedFS is an optional filesystem containing starter kits.
// Usually set from main.go via go:embed.
var EmbeddedFS embed.FS

// EmbeddedPath is the path within EmbeddedFS where starters are located.
var EmbeddedPath = "starters"

// Starter represents a discovered starter kit.
type Starter struct {
	Name         string // Derived from directory name
	Description  string // Parsed from protocol.yaml
	ProtocolPath string // Path to protocol.yaml
	TemplatesDir string // Path to templates directory
	SkillsDir    string // Path to skills directory (optional)
	DefaultsPath string // Path to defaults.yaml (optional)
	SourceFS     fs.FS  // FS containing the starter (nil if local disk)
	IsBuiltin    bool   // True if loaded from embedded FS
}

// Defaults represents optional configuration defaults from a starter.
type Defaults struct {
	Language    string            `yaml:"language,omitempty"`
	Framework   string            `yaml:"framework,omitempty"`
	CustomVars  map[string]string `yaml:"custom_vars,omitempty"`
	Constraints map[string]string `yaml:"constraints,omitempty"`
}

// Discover scans a local base path for valid starter kits.
func Discover(basePath string) ([]Starter, error) {
	return DiscoverFromFS(os.DirFS(basePath), ".")
}

// DiscoverFromFS scans a filesystem for valid starter kits.
func DiscoverFromFS(sys fs.FS, basePath string) ([]Starter, error) {
	entries, err := fs.ReadDir(sys, basePath)
	if err != nil {
		return nil, fmt.Errorf("reading starter base path: %w", err)
	}

	var starters []Starter
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		starterDir := filepath.Join(basePath, name)

		// Check for required protocol.yaml
		protocolPath := filepath.Join(starterDir, "protocol.yaml")
		data, err := fs.ReadFile(sys, protocolPath)
		if err != nil {
			continue
		}

		// Check for required templates directory
		templatesDir := filepath.Join(starterDir, "templates")
		info, err := fs.Stat(sys, templatesDir)
		if err != nil || !info.IsDir() {
			continue
		}

		starter := Starter{
			Name:         name,
			ProtocolPath: protocolPath,
			TemplatesDir: templatesDir,
			SourceFS:     sys,
		}

		// Basic YAML parsing to get description
		var meta struct {
			Name        string `yaml:"name"`
			Description string `yaml:"description"`
		}
		if err := yaml.Unmarshal(data, &meta); err == nil {
			if meta.Description != "" {
				starter.Description = meta.Description
			} else if meta.Name != "" && meta.Name != name {
				starter.Description = meta.Name
			}
		}

		// Check for optional skills directory
		skillsDir := filepath.Join(starterDir, "skills")
		if info, err := fs.Stat(sys, skillsDir); err == nil && info.IsDir() {
			starter.SkillsDir = skillsDir
		}

		// Check for optional defaults.yaml
		defaultsPath := filepath.Join(starterDir, "defaults.yaml")
		if _, err := fs.Stat(sys, defaultsPath); err == nil {
			starter.DefaultsPath = defaultsPath
		}

		starters = append(starters, starter)
	}

	return starters, nil
}

// List returns available starter names.
// It combines starters found on disk with embedded ones.
func List() ([]Starter, error) {
	var allStarters []Starter
	seen := make(map[string]bool)

	// 1. Try local disk (takes precedence for developers)
	candidates := findStarterDirs()
	for _, dir := range candidates {
		local, err := Discover(dir)
		if err == nil {
			for _, s := range local {
				if !seen[s.Name] {
					allStarters = append(allStarters, s)
					seen[s.Name] = true
				}
			}
		}
	}

	// 2. Try embedded FS
	embedded, err := DiscoverFromFS(EmbeddedFS, EmbeddedPath)
	if err == nil {
		for _, s := range embedded {
			if !seen[s.Name] {
				s.IsBuiltin = true
				allStarters = append(allStarters, s)
				seen[s.Name] = true
			}
		}
	}

	if len(allStarters) == 0 {
		return nil, fmt.Errorf("no starters found; ensure starters/ directory exists and contains valid workflows")
	}

	// Sort by name for consistent output
	sort.Slice(allStarters, func(i, j int) bool {
		return allStarters[i].Name < allStarters[j].Name
	})

	return allStarters, nil
}

// findStarterDirs attempts to locate starter directories on disk.
func findStarterDirs() []string {
	var dirs []string
	if wd, err := os.Getwd(); err == nil {
		bases := []string{wd, filepath.Join(wd, ".."), filepath.Join(wd, "..", "..")}
		for _, base := range bases {
			s := filepath.Join(base, "starters")
			if info, err := os.Stat(s); err == nil && info.IsDir() {
				dirs = append(dirs, s)
			}
		}
	}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		bases := []string{exeDir, filepath.Join(exeDir, ".."), filepath.Join(exeDir, "..", "share", "specfirst")}
		for _, base := range bases {
			s := filepath.Join(base, "starters")
			if info, err := os.Stat(s); err == nil && info.IsDir() {
				dirs = append(dirs, s)
			}
		}
	}
	return dirs
}

// Apply installs a starter kit to the current workspace.
// If force is false, existing files are not overwritten.
// If updateConfig is true, config.yaml protocol is updated.
func Apply(name string, force bool, updateConfig bool) error {
	starters, err := List()
	if err != nil {
		return err
	}

	var starter *Starter
	for i := range starters {
		if starters[i].Name == name {
			starter = &starters[i]
			break
		}
	}
	if starter == nil {
		return fmt.Errorf("starter %q not found", name)
	}

	// Ensure workspace directories exist
	if err := ensureDir(repository.ProtocolsPath()); err != nil {
		return err
	}
	// Namespace templates: .specfirst/templates/<starter-name>/
	templatesDir := repository.TemplatesPath(name)
	if err := ensureDir(templatesDir); err != nil {
		return err
	}
	if starter.SkillsDir != "" {
		if err := ensureDir(repository.SkillsPath()); err != nil {
			return err
		}
	}

	// 1. Rewrite and save protocol
	// Prefix all template paths in the protocol with the starter name
	protocolData, err := rewriteProtocolTemplates(starter.SourceFS, starter.ProtocolPath, name)
	if err != nil {
		return fmt.Errorf("rewriting protocol: %w", err)
	}

	destProtocol := repository.ProtocolsPath(name + ".yaml")
	if !force {
		if _, err := os.Stat(destProtocol); err == nil {
			// Protocol exists, don't overwrite
		} else {
			if err := os.WriteFile(destProtocol, protocolData, 0644); err != nil {
				return fmt.Errorf("writing protocol: %w", err)
			}
		}
	} else {
		if err := os.WriteFile(destProtocol, protocolData, 0644); err != nil {
			return fmt.Errorf("writing protocol: %w", err)
		}
	}

	// 2. Copy templates to namespaced directory
	if err := copyDirFromFS(starter.SourceFS, starter.TemplatesDir, templatesDir, force); err != nil {
		return fmt.Errorf("copying templates: %w", err)
	}

	// 3. Copy skills if present
	if starter.SkillsDir != "" {
		if err := copyDirFromFS(starter.SourceFS, starter.SkillsDir, repository.SkillsPath(), force); err != nil {
			return fmt.Errorf("copying skills: %w", err)
		}
	}

	// 4. Update config if requested
	if updateConfig {
		if err := updateConfigProtocol(name, starter.SourceFS, starter.DefaultsPath, force); err != nil {
			return fmt.Errorf("updating config: %w", err)
		}
	}

	return nil
}

// rewriteProtocolTemplates prefixes all template paths in the protocol with the starter name.
func rewriteProtocolTemplates(sys fs.FS, protocolPath string, starterName string) ([]byte, error) {
	data, err := fs.ReadFile(sys, protocolPath)
	if err != nil {
		return nil, err
	}

	var protocol map[string]interface{}
	if err := yaml.Unmarshal(data, &protocol); err != nil {
		return nil, err
	}

	stages, ok := protocol["stages"].([]interface{})
	if !ok {
		return data, nil
	}

	for _, s := range stages {
		stage, ok := s.(map[string]interface{})
		if !ok {
			continue
		}
		if template, ok := stage["template"].(string); ok && template != "" {
			// Prefix with starter name using forward slashes for protocol portability
			stage["template"] = starterName + "/" + template
		}
	}

	return yaml.Marshal(protocol)
}

// ensureDir creates a directory if it doesn't exist.
func ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// copyFileFromFS copies a file from an fs.FS to local disk.
func copyFileFromFS(sys fs.FS, src, dst string, force bool) error {
	if !force {
		if _, err := os.Stat(dst); err == nil {
			return nil // File exists, don't overwrite
		}
	}

	srcFile, err := sys.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// copyDirFromFS copies all files from a directory in fs.FS to local disk.
func copyDirFromFS(sys fs.FS, srcDir, dstDir string, force bool) error {
	entries, err := fs.ReadDir(sys, srcDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(dstDir, entry.Name())

		if err := copyFileFromFS(sys, srcPath, dstPath, force); err != nil {
			return fmt.Errorf("copying %s: %w", entry.Name(), err)
		}
	}

	return nil
}

// updateConfigProtocol updates config.yaml with the starter protocol.
func updateConfigProtocol(protocolName string, sys fs.FS, defaultsPath string, force bool) error {
	configPath := repository.ConfigPath()

	// Read existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}

	// Update protocol
	config["protocol"] = protocolName

	// Apply defaults if present
	if defaultsPath != "" {
		defaults, err := parseDefaultsFromFS(sys, defaultsPath)
		if err != nil {
			return err
		}
		applyDefaults(config, defaults, force)
	}

	// Write back
	out, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, out, 0644)
}

// parseDefaultsFromFS reads a defaults.yaml file from an fs.FS.
func parseDefaultsFromFS(sys fs.FS, path string) (*Defaults, error) {
	data, err := fs.ReadFile(sys, path)
	if err != nil {
		return nil, err
	}

	var defaults Defaults
	if err := yaml.Unmarshal(data, &defaults); err != nil {
		return nil, err
	}

	return &defaults, nil
}

// applyDefaults merges defaults into config.
// If force is false, only missing keys are set.
func applyDefaults(config map[string]interface{}, defaults *Defaults, force bool) {
	if defaults == nil {
		return
	}

	setIfMissing := func(key, value string) {
		if value == "" {
			return
		}
		if force {
			config[key] = value
		} else if _, exists := config[key]; !exists || config[key] == "" {
			config[key] = value
		}
	}

	setIfMissing("language", defaults.Language)
	setIfMissing("framework", defaults.Framework)

	// Merge custom_vars
	if len(defaults.CustomVars) > 0 {
		existing, ok := config["custom_vars"].(map[string]interface{})
		if !ok {
			existing = make(map[string]interface{})
		}
		for k, v := range defaults.CustomVars {
			if force {
				existing[k] = v
			} else if _, exists := existing[k]; !exists {
				existing[k] = v
			}
		}
		config["custom_vars"] = existing
	}

	// Merge constraints
	if len(defaults.Constraints) > 0 {
		existing, ok := config["constraints"].(map[string]interface{})
		if !ok {
			existing = make(map[string]interface{})
		}
		for k, v := range defaults.Constraints {
			if force {
				existing[k] = v
			} else if _, exists := existing[k]; !exists {
				existing[k] = v
			}
		}
		config["constraints"] = existing
	}
}
