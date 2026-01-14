package cmd

import (
	"strings"
	"testing"

	"specfirst/internal/bundle"
)

func TestRenderBundleBody_RawOmitsHeadings(t *testing.T) {
	out := renderBundleBody("stage", "PROMPT", []bundle.File{{Path: "a.txt", Content: "A"}}, true)
	if strings.Contains(out, "## Prompt") || strings.Contains(out, "## Files") {
		t.Fatalf("expected no headings, got: %q", out)
	}
	if !strings.Contains(out, "<prompt stage=\"stage\">\nPROMPT\n</prompt>") {
		t.Fatalf("expected prompt block, got: %q", out)
	}
	if !strings.Contains(out, "<file path=\"a.txt\">\nA\n</file>") {
		t.Fatalf("expected file block, got: %q", out)
	}
}

func TestHeredocDelimiter_AvoidsCollisions(t *testing.T) {
	body := "hello\nSPECFIRST_BUNDLE_EOF\nworld"
	d := heredocDelimiter(body)
	if d == "SPECFIRST_BUNDLE_EOF" {
		t.Fatalf("expected a non-colliding delimiter")
	}
}
