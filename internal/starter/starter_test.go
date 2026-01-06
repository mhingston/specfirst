package starter

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestRewriteProtocolTemplates(t *testing.T) {
	tmpDir := t.TempDir()
	protocolPath := filepath.Join(tmpDir, "protocol.yaml")

	content := `name: test-starter
stages:
  - name: stage1
    template: template1.md
  - name: stage2
    template: template2.md
`
	if err := os.WriteFile(protocolPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	sys := os.DirFS(tmpDir)
	rewritten, err := rewriteProtocolTemplates(sys, "protocol.yaml", "test-starter")
	if err != nil {
		t.Fatalf("rewriteProtocolTemplates failed: %v", err)
	}

	var protocol map[string]interface{}
	if err := yaml.Unmarshal(rewritten, &protocol); err != nil {
		t.Fatal(err)
	}

	stages := protocol["stages"].([]interface{})
	s1 := stages[0].(map[string]interface{})
	if s1["template"] != "test-starter/template1.md" {
		t.Errorf("expected test-starter/template1.md, got %v", s1["template"])
	}
}

func TestDiscover(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Create a valid starter
	validStarterDir := filepath.Join(tmpDir, "valid-starter")
	if err := os.MkdirAll(filepath.Join(validStarterDir, "templates"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(validStarterDir, "protocol.yaml"), []byte("name: valid\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create an invalid starter (missing templates)
	invalidStarterDir := filepath.Join(tmpDir, "invalid-starter")
	if err := os.MkdirAll(invalidStarterDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(invalidStarterDir, "protocol.yaml"), []byte("name: invalid\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a starter with skills
	skillsStarterDir := filepath.Join(tmpDir, "with-skills")
	if err := os.MkdirAll(filepath.Join(skillsStarterDir, "templates"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(skillsStarterDir, "skills"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillsStarterDir, "protocol.yaml"), []byte("name: with-skills\n"), 0644); err != nil {
		t.Fatal(err)
	}

	starters, err := Discover(tmpDir)
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}

	if len(starters) != 2 {
		t.Errorf("expected 2 starters, got %d", len(starters))
	}

	// Check valid-starter
	found := false
	for _, s := range starters {
		if s.Name == "valid-starter" {
			found = true
			if s.SkillsDir != "" {
				t.Error("valid-starter should not have skills")
			}
		}
	}
	if !found {
		t.Error("valid-starter not found")
	}

	// Check with-skills
	found = false
	for _, s := range starters {
		if s.Name == "with-skills" {
			found = true
			if s.SkillsDir == "" {
				t.Error("with-skills should have skills directory")
			}
		}
	}
	if !found {
		t.Error("with-skills not found")
	}
}

func TestCopyFileFromFS(t *testing.T) {
	tmpDir := t.TempDir()

	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	srcPath := "test.txt"
	dstPath := filepath.Join(tmpDir, "dst.txt")

	// Create source file
	content := "test content"
	if err := os.WriteFile(filepath.Join(srcDir, srcPath), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	fsys := os.DirFS(srcDir)

	// Test basic copy
	if err := copyFileFromFS(fsys, srcPath, dstPath, false); err != nil {
		t.Fatalf("copyFileFromFS failed: %v", err)
	}

	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != content {
		t.Errorf("got %q, want %q", string(got), content)
	}
}

func TestCopyFileFromFSNoOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}

	srcPath := "src.txt"
	dstPath := filepath.Join(tmpDir, "dst.txt")

	// Create source and destination files
	if err := os.WriteFile(filepath.Join(srcDir, srcPath), []byte("new content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dstPath, []byte("original content"), 0644); err != nil {
		t.Fatal(err)
	}

	fsys := os.DirFS(srcDir)

	// Copy without force should not overwrite
	if err := copyFileFromFS(fsys, srcPath, dstPath, false); err != nil {
		t.Fatalf("copyFileFromFS failed: %v", err)
	}

	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "original content" {
		t.Errorf("file was overwritten, got %q", string(got))
	}
}

func TestCopyFileFromFSWithForce(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}

	srcPath := "src.txt"
	dstPath := filepath.Join(tmpDir, "dst.txt")

	// Create source and destination files
	if err := os.WriteFile(filepath.Join(srcDir, srcPath), []byte("new content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dstPath, []byte("original content"), 0644); err != nil {
		t.Fatal(err)
	}

	fsys := os.DirFS(srcDir)

	// Copy with force should overwrite
	if err := copyFileFromFS(fsys, srcPath, dstPath, true); err != nil {
		t.Fatalf("copyFileFromFS failed: %v", err)
	}

	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new content" {
		t.Errorf("file was not overwritten, got %q", string(got))
	}
}

func TestParseDefaultsFromFS(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	defaultsPath := "defaults.yaml"

	content := `language: Go
framework: gin
custom_vars:
  api_version: v1
constraints:
  max_response_time: 100ms
`
	if err := os.WriteFile(filepath.Join(srcDir, defaultsPath), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	fsys := os.DirFS(srcDir)

	defaults, err := parseDefaultsFromFS(fsys, defaultsPath)
	if err != nil {
		t.Fatalf("parseDefaultsFromFS failed: %v", err)
	}

	if defaults.Language != "Go" {
		t.Errorf("language: got %q, want Go", defaults.Language)
	}
	if defaults.Framework != "gin" {
		t.Errorf("framework: got %q, want gin", defaults.Framework)
	}
	if defaults.CustomVars["api_version"] != "v1" {
		t.Errorf("custom_vars.api_version: got %q, want v1", defaults.CustomVars["api_version"])
	}
	if defaults.Constraints["max_response_time"] != "100ms" {
		t.Errorf("constraints.max_response_time: got %q, want 100ms", defaults.Constraints["max_response_time"])
	}
}

func TestApplyDefaults(t *testing.T) {
	config := map[string]interface{}{
		"language": "",
	}

	defaults := &Defaults{
		Language:  "TypeScript",
		Framework: "Next.js",
	}

	applyDefaults(config, defaults, false)

	if config["language"] != "TypeScript" {
		t.Errorf("language: got %q, want TypeScript", config["language"])
	}
	if config["framework"] != "Next.js" {
		t.Errorf("framework: got %q, want Next.js", config["framework"])
	}
}

func TestApplyDefaultsNoOverwrite(t *testing.T) {
	config := map[string]interface{}{
		"language": "Go",
	}

	defaults := &Defaults{
		Language: "TypeScript",
	}

	applyDefaults(config, defaults, false)

	// Should not overwrite existing value
	if config["language"] != "Go" {
		t.Errorf("language: got %q, want Go (should not be overwritten)", config["language"])
	}
}
