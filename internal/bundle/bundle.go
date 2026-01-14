package bundle

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"specfirst/internal/repository"
)

type Options struct {
	IncludePatterns []string
	ExcludePatterns []string

	MaxFiles        int
	MaxTotalBytes   int64
	MaxFileBytes    int64
	DefaultExcludes bool
}

type File struct {
	Path    string // slash-separated, project-relative
	Content string
	Bytes   int64
}

type Report struct {
	IncludedFiles int
	IncludedBytes int64

	SkippedByExclude int
	SkippedTooLarge  int
	SkippedOverLimit int
	MissingLiterals  []string
}

var ErrNoFilesSelected = errors.New("no files selected")

func Collect(opts Options) ([]File, Report, error) {
	if opts.MaxFiles <= 0 {
		opts.MaxFiles = 50
	}
	if opts.MaxTotalBytes <= 0 {
		opts.MaxTotalBytes = 250_000
	}
	if opts.MaxFileBytes <= 0 {
		opts.MaxFileBytes = 100_000
	}
	if opts.DefaultExcludes {
		opts.ExcludePatterns = append(defaultExcludes(), opts.ExcludePatterns...)
	}

	root := repository.BaseDir()
	includes := normalizePatterns(opts.IncludePatterns)
	excludes := normalizePatterns(opts.ExcludePatterns)

	literalWanted := literalPatterns(includes)
	literalFound := make(map[string]bool)

	var report Report

	candidates := make([]string, 0, 128)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Fast-path skip common heavy dirs.
			name := d.Name()
			if name == ".git" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		if len(includes) == 0 {
			return nil
		}
		if !matchesAny(includes, rel) {
			return nil
		}
		if matchesAny(excludes, rel) {
			report.SkippedByExclude++
			return nil
		}

		// Mark literal hit.
		if _, ok := literalWanted[rel]; ok {
			literalFound[rel] = true
		}

		candidates = append(candidates, rel)
		return nil
	})
	if err != nil {
		return nil, Report{}, err
	}

	for lit := range literalWanted {
		if !literalFound[lit] {
			// If it's a literal, check it exists even if it lives under an excluded dir.
			abs := filepath.Join(root, filepath.FromSlash(lit))
			if _, err := os.Stat(abs); err == nil {
				candidates = append(candidates, lit)
				literalFound[lit] = true
			} else {
				// Keep track for reporting.
			}
		}
	}

	sort.Strings(candidates)
	candidates = uniqueStrings(candidates)

	if len(candidates) == 0 {
		return nil, Report{}, ErrNoFilesSelected
	}

	for lit := range literalWanted {
		if !literalFound[lit] {
			report.MissingLiterals = append(report.MissingLiterals, lit)
		}
	}
	sort.Strings(report.MissingLiterals)

	selected := make([]File, 0, minInt(opts.MaxFiles, len(candidates)))
	var totalBytes int64
	for _, rel := range candidates {
		if len(selected) >= opts.MaxFiles {
			report.SkippedOverLimit++
			continue
		}

		abs := filepath.Join(root, filepath.FromSlash(rel))
		info, err := os.Stat(abs)
		if err != nil {
			continue
		}
		size := info.Size()
		if size > opts.MaxFileBytes {
			report.SkippedTooLarge++
			continue
		}
		if totalBytes+size > opts.MaxTotalBytes {
			report.SkippedOverLimit++
			continue
		}

		b, err := os.ReadFile(abs)
		if err != nil {
			continue
		}

		selected = append(selected, File{Path: rel, Content: string(b), Bytes: int64(len(b))})
		totalBytes += int64(len(b))
	}

	report.IncludedFiles = len(selected)
	report.IncludedBytes = totalBytes

	if report.IncludedFiles == 0 {
		return nil, report, ErrNoFilesSelected
	}

	return selected, report, nil
}

func defaultExcludes() []string {
	return []string{
		".git/**",
		".specfirst/**",
		"dist/**",
		"tmp/**",
		"node_modules/**",
	}
}

func normalizePatterns(patterns []string) []string {
	out := make([]string, 0, len(patterns))
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		p = strings.TrimPrefix(p, "./")
		p = strings.ReplaceAll(p, "\\", "/")
		out = append(out, p)
	}
	return out
}

func literalPatterns(patterns []string) map[string]struct{} {
	literals := make(map[string]struct{})
	for _, p := range patterns {
		if !strings.ContainsAny(p, "*?[") && !strings.Contains(p, "**") {
			literals[p] = struct{}{}
		}
	}
	return literals
}

func matchesAny(patterns []string, rel string) bool {
	for _, p := range patterns {
		ok, err := matchGlob(p, rel)
		if err != nil {
			continue
		}
		if ok {
			return true
		}
	}
	return false
}

func uniqueStrings(items []string) []string {
	if len(items) == 0 {
		return items
	}
	out := []string{items[0]}
	for i := 1; i < len(items); i++ {
		if items[i] != items[i-1] {
			out = append(out, items[i])
		}
	}
	return out
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// matchGlob matches rel paths (slash-separated) against a glob pattern.
// Supports ** for recursive matching.
func matchGlob(pattern string, rel string) (bool, error) {
	pattern = strings.TrimPrefix(pattern, "./")
	pattern = strings.ReplaceAll(pattern, "\\", "/")
	rel = strings.TrimPrefix(rel, "./")
	rel = strings.ReplaceAll(rel, "\\", "/")

	if pattern == "**" {
		return true, nil
	}

	pParts := splitKeepEmpty(pattern)
	rParts := splitKeepEmpty(rel)
	return matchParts(pParts, rParts)
}

func splitKeepEmpty(p string) []string {
	// Preserve empty parts? Not needed; treat consecutive slashes as one.
	p = strings.Trim(p, "/")
	if p == "" {
		return []string{}
	}
	return strings.Split(p, "/")
}

func matchParts(patternParts []string, pathParts []string) (bool, error) {
	if len(patternParts) == 0 {
		return len(pathParts) == 0, nil
	}

	head := patternParts[0]
	if head == "**" {
		// ** can match zero or more segments.
		// Try zero segments.
		ok, err := matchParts(patternParts[1:], pathParts)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
		// Try consuming one segment.
		if len(pathParts) == 0 {
			return false, nil
		}
		return matchParts(patternParts, pathParts[1:])
	}

	if len(pathParts) == 0 {
		return false, nil
	}

	ok, err := filepath.Match(head, pathParts[0])
	if err != nil {
		return false, fmt.Errorf("invalid glob %q: %w", head, err)
	}
	if !ok {
		return false, nil
	}

	return matchParts(patternParts[1:], pathParts[1:])
}
