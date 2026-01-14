package templating

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"specfirst/internal/repository"
)

func TestRender_TemplateFunctions(t *testing.T) {
	root := t.TempDir()
	repository.SetRootDir(root)
	t.Cleanup(repository.ResetRootDir)

	// Skills are searched first.
	if err := os.MkdirAll(filepath.Join(root, ".specfirst", "skills"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".specfirst", "skills", "foo.txt"), []byte("SKILL"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "foo.txt"), []byte("ROOT"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(root, "inputs"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "inputs", "bar.txt"), []byte("BAR"), 0644); err != nil {
		t.Fatal(err)
	}

	tplPath := filepath.Join(root, "tpl.md")
	tpl := strings.Join([]string{
		"foo={{ readFile \"foo.txt\" }}",
		"bar={{ readFile \"inputs/bar.txt\" }}",
		"out={{ join .Outputs \",\" }}",
		"u={{ upper \"hi\" }}",
		"l={{ lower \"HI\" }}",
		"",
	}, "\n")
	if err := os.WriteFile(tplPath, []byte(tpl), 0644); err != nil {
		t.Fatal(err)
	}

	out, err := Render(tplPath, Data{Outputs: []string{"a", "b"}})
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	if !strings.Contains(out, "foo=SKILL") {
		t.Fatalf("expected skills file to win, got: %q", out)
	}
	if !strings.Contains(out, "bar=BAR") {
		t.Fatalf("expected project-root file include, got: %q", out)
	}
	if !strings.Contains(out, "out=a,b") {
		t.Fatalf("expected join helper, got: %q", out)
	}
	if !strings.Contains(out, "u=HI") || !strings.Contains(out, "l=hi") {
		t.Fatalf("expected upper/lower helpers, got: %q", out)
	}
}

func TestRender_ReadFileRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	repository.SetRootDir(root)
	t.Cleanup(repository.ResetRootDir)

	tplPath := filepath.Join(root, "tpl.md")
	if err := os.WriteFile(tplPath, []byte("{{ readFile \"../secrets.txt\" }}"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Render(tplPath, Data{})
	if err == nil {
		t.Fatalf("expected error")
	}
}
