package templating

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"specfirst/internal/repository"
)

type Input struct {
	Name    string
	Content string
}

type Data struct {
	StageName   string
	ProjectName string
	Inputs      []Input
	Outputs     []string
	Intent      string
	Language    string
	Framework   string
	CustomVars  map[string]string
	Constraints map[string]string

	// Protocol v2 fields
	StageType      string
	Prompt         any
	OutputContract any
	Epistemics     any
}

func templateFuncMap() template.FuncMap {
	return template.FuncMap{
		"join": func(items []string, sep string) string {
			return strings.Join(items, sep)
		},
		"upper":    strings.ToUpper,
		"lower":    strings.ToLower,
		"readFile": readFile,
	}
}

func sanitizeRelPath(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("readFile: path is empty")
	}

	normalized := strings.ReplaceAll(value, "\\", "/")
	clean := filepath.Clean(filepath.FromSlash(normalized))

	if filepath.IsAbs(clean) || isWindowsAbs(normalized) {
		return "", fmt.Errorf("readFile: absolute paths are not allowed: %s", value)
	}
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("readFile: invalid path: %s", value)
	}
	for _, part := range strings.Split(filepath.ToSlash(clean), "/") {
		if part == ".." {
			return "", fmt.Errorf("readFile: invalid path: %s", value)
		}
	}

	return filepath.ToSlash(clean), nil
}

func isWindowsAbs(value string) bool {
	if len(value) >= 3 && value[1] == ':' && (value[2] == '/' || value[2] == '\\') {
		return true
	}
	return strings.HasPrefix(value, "//") || strings.HasPrefix(value, `\\`)
}

func tryReadProjectRel(rel string) (string, bool, error) {
	abs := filepath.Join(repository.BaseDir(), filepath.FromSlash(rel))
	data, err := os.ReadFile(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("readFile: reading %s: %w", rel, err)
	}
	return string(data), true, nil
}

func readFile(path string) (string, error) {
	rel, err := sanitizeRelPath(path)
	if err != nil {
		return "", err
	}

	skillsRel := filepath.ToSlash(filepath.Join(repository.SpecDir, repository.SkillsDir, filepath.FromSlash(rel)))
	if content, ok, err := tryReadProjectRel(skillsRel); err != nil {
		return "", err
	} else if ok {
		return content, nil
	}

	if content, ok, err := tryReadProjectRel(rel); err != nil {
		return "", err
	} else if ok {
		return content, nil
	}

	return "", fmt.Errorf("readFile: file not found: %s (searched %s and %s)", rel, skillsRel, rel)
}

// Render renders a template file with the given data.
func Render(templatePath string, data Data) (string, error) {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("reading template %s: %w", templatePath, err)
	}

	tmpl, err := template.New(templatePath).Funcs(templateFuncMap()).Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("parsing template %s: %w", templatePath, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing template %s: %w", templatePath, err)
	}

	return buf.String(), nil
}

// RenderInline renders an inline template string.
func RenderInline(tmpl string, data any) (string, error) {
	parsed, err := template.New("inline").Funcs(templateFuncMap()).Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	if err := parsed.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
