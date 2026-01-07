package repository

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"specfirst/internal/utils"
)

// ArtifactPathForInput resolves an input artifact path.
func ArtifactPathForInput(filename string, priorityStages []string, stageIDs []string) (string, error) {
	if filename == "" {
		return "", fmt.Errorf("invalid input artifact path: %s", filename)
	}

	clean := filepath.Clean(filename)
	if filepath.IsAbs(clean) || clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid input artifact path: %s", filename)
	}
	for _, part := range strings.Split(filepath.ToSlash(clean), "/") {
		if part == ".." {
			return "", fmt.Errorf("invalid input artifact path: %s", filename)
		}
	}

	// Support stage-qualified paths like "stage-id/filename" when stage-id matches a known stage.
	stageQualified := false
	if strings.Contains(clean, "/") || strings.Contains(clean, string(os.PathSeparator)) {
		first := strings.SplitN(filepath.ToSlash(clean), "/", 2)[0]
		for _, id := range stageIDs {
			if id == first {
				stageQualified = true
				break
			}
		}
	}
	if stageQualified {
		path := ArtifactsPath(clean)
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return "", fmt.Errorf("missing input artifact: %s", filename)
			}
			return "", err
		}
		return path, nil
	}

	// Priority Check Level 1: Check in the priority stages first (explicit dependencies)
	for _, stageID := range priorityStages {
		path := ArtifactsPath(stageID, clean)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// Glob-based resolution for unqualified filenames (fallback)
	pattern := ArtifactsPath("*", clean)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("missing input artifact: %s", filename)
	}
	if len(matches) > 1 {
		// Extract stage names from paths for helpful suggestion
		stages := make([]string, 0, len(matches))
		for _, match := range matches {
			rel, _ := filepath.Rel(ArtifactsPath(), match)
			if parts := strings.SplitN(rel, string(os.PathSeparator), 2); len(parts) > 0 {
				stages = append(stages, parts[0])
			}
		}
		sort.Strings(stages) // Alphabetical for consistent, deterministic output
		return "", fmt.Errorf("ambiguous input artifact %q found in multiple stages: %v\nHint: use a stage-qualified path like %q", filename, stages, stages[0]+"/"+clean)
	}
	return matches[0], nil
}

func ArtifactRelFromState(value string) (string, error) {
	if value == "" {
		return "", fmt.Errorf("invalid artifact path: %q", value)
	}
	normalized := strings.ReplaceAll(value, "\\", "/")
	clean := filepath.Clean(filepath.FromSlash(normalized))
	if filepath.IsAbs(clean) || isWindowsAbs(normalized) {
		rel, ok := relFromArtifactsPath(clean)
		if !ok {
			return "", fmt.Errorf("artifact path is outside artifacts dir: %s", value)
		}
		return filepath.ToSlash(rel), nil
	}
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid artifact path: %s", value)
	}
	return filepath.ToSlash(clean), nil
}

func ArtifactAbsFromState(value string) (string, error) {
	rel, err := ArtifactRelFromState(value)
	if err != nil {
		return "", err
	}
	return filepath.Join(ArtifactsPath(), filepath.FromSlash(rel)), nil
}

func isWindowsAbs(value string) bool {
	if len(value) >= 3 && value[1] == ':' && value[2] == '/' {
		return true
	}
	return strings.HasPrefix(value, "//")
}

func relFromArtifactsPath(abs string) (string, bool) {
	clean := filepath.Clean(abs)
	parts := strings.Split(clean, string(os.PathSeparator))
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == ArtifactsDir {
			if i+1 >= len(parts) {
				return "", false
			}
			rel := filepath.Join(parts[i+1:]...)
			if rel == "" || rel == "." {
				return "", false
			}
			return rel, true
		}
	}
	return "", false
}

// NormalizeMatchPath normalizes a path for pattern matching.
func NormalizeMatchPath(value string) string {
	normalized := filepath.ToSlash(value)
	normalized = strings.ReplaceAll(normalized, "\\", "/")
	return strings.TrimPrefix(normalized, "./")
}

// MatchOutputPattern checks if a file matches an output pattern.
func MatchOutputPattern(pattern string, file string) bool {
	if pattern == "" || file == "" {
		return false
	}
	cleanPattern := NormalizeMatchPath(pattern)
	cleanFile := NormalizeMatchPath(file)
	// Try matching against relative path and basename
	candidates := []string{cleanFile, path.Base(cleanFile)}
	for _, candidate := range candidates {
		ok, err := path.Match(cleanPattern, candidate)
		if err != nil {
			// Sev1 Fix: Don't silently ignore invalid patterns.
			// Since we cannot easily return error from here without cascading changes,
			// logging is the best immediate safety net.
			fmt.Fprintf(os.Stderr, "Warning: invalid output pattern %q in protocol: %v\n", pattern, err)
			return false
		}
		if ok {
			return true
		}
	}
	return cleanFile == cleanPattern
}

// WriteOutput writes content to a file, ensuring the directory exists.
func WriteOutput(path string, content string) error {
	if err := utils.EnsureDir(filepath.Dir(path)); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// ProjectRelPath returns the path relative to the project root.
func ProjectRelPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path is empty")
	}

	clean := filepath.Clean(path)

	// Determine project root
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	root, err := FindProjectRoot(wd)
	if err != nil {
		return "", err
	}

	// Resolve absolute path
	abs := clean
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(wd, clean)
	}

	// Evaluate symlinks for canonical path comparison
	// We must resolve the root to its canonical form
	if rootEval, err := filepath.EvalSymlinks(root); err == nil {
		root = rootEval
	}

	// We must also resolve abs to its canonical form as much as possible
	// Walk up until we find a directory that exists
	base := abs
	suffix := ""
	for {
		if _, err := os.Stat(base); err == nil {
			// Found existing path
			break
		}
		parent := filepath.Dir(base)
		if parent == base {
			break
		}
		suffix = filepath.Join(filepath.Base(base), suffix)
		base = parent
	}

	if baseEval, err := filepath.EvalSymlinks(base); err == nil {
		if suffix != "" {
			abs = filepath.Join(baseEval, suffix)
		} else {
			abs = baseEval
		}
	}

	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return "", err
	}

	// Safety checks
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("path escapes project root: %s", path)
	}
	if rel == "." {
		return "", fmt.Errorf("path resolves to project root")
	}

	return rel, nil
}

// ResolveOutputPath resolves an output path relative to the project root.
func ResolveOutputPath(output string) (string, error) {
	rel, err := ProjectRelPath(output)
	if err != nil {
		return "", err
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	root, err := FindProjectRoot(wd)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, rel), nil
}
