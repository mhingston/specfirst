package bundle

import (
	"os"
	"path/filepath"
	"testing"

	"specfirst/internal/repository"
)

func TestCollect_SupportsDoubleStar(t *testing.T) {
	root := t.TempDir()
	repository.SetRootDir(root)
	t.Cleanup(repository.ResetRootDir)

	if err := os.MkdirAll(filepath.Join(root, "src", "a"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "a", "one.txt"), []byte("1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "two.txt"), []byte("2"), 0644); err != nil {
		t.Fatal(err)
	}

	files, report, err := Collect(Options{
		IncludePatterns: []string{"src/**"},
		ExcludePatterns: []string{"src/a/**"},
		MaxFiles:        50,
		MaxTotalBytes:   250_000,
		MaxFileBytes:    100_000,
		DefaultExcludes: false,
	})
	if err != nil {
		t.Fatalf("Collect() error: %v", err)
	}
	if report.IncludedFiles != 1 {
		t.Fatalf("expected 1 file, got %d", report.IncludedFiles)
	}
	if len(files) != 1 || files[0].Path != "src/two.txt" {
		t.Fatalf("unexpected files: %+v", files)
	}
}

func TestCollect_RespectsMaxTotalBytes(t *testing.T) {
	root := t.TempDir()
	repository.SetRootDir(root)
	t.Cleanup(repository.ResetRootDir)

	if err := os.MkdirAll(filepath.Join(root, "src"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "a.txt"), []byte("aaaaa"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "b.txt"), []byte("bbbbb"), 0644); err != nil {
		t.Fatal(err)
	}

	files, report, err := Collect(Options{
		IncludePatterns: []string{"src/**"},
		MaxFiles:        50,
		MaxTotalBytes:   6,
		MaxFileBytes:    100_000,
	})
	if err != nil {
		t.Fatalf("Collect() error: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file included, got %d", len(files))
	}
	if report.IncludedBytes > 6 {
		t.Fatalf("expected <= 6 bytes, got %d", report.IncludedBytes)
	}
}
