package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiffCommand(t *testing.T) {
	t.Run("generates diff prompt", func(t *testing.T) {
		oldFile := filepath.Join(t.TempDir(), "old.md")
		newFile := filepath.Join(t.TempDir(), "new.md")

		if err := os.WriteFile(oldFile, []byte("# Old\n\n- Feature A"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(newFile, []byte("# New\n\n- Feature B"), 0644); err != nil {
			t.Fatal(err)
		}

		prompt := generateDiffPrompt("# Old\n\n- Feature A", "# New\n\n- Feature B", oldFile, newFile)

		if !strings.Contains(prompt, "Behavioral differences") {
			t.Error("expected prompt to contain 'Behavioral differences'")
		}
		if !strings.Contains(prompt, "Backward compatibility risks") {
			t.Error("expected prompt to contain 'Backward compatibility risks'")
		}
		if !strings.Contains(prompt, "Previous Specification") {
			t.Error("expected prompt to contain 'Previous Specification'")
		}
		if !strings.Contains(prompt, "Proposed Specification") {
			t.Error("expected prompt to contain 'Proposed Specification'")
		}
	})

	t.Run("command requires two args", func(t *testing.T) {
		var buf bytes.Buffer
		diffCmd.SetOut(&buf)
		diffCmd.SetErr(&buf)
		diffCmd.SetArgs([]string{})

		err := diffCmd.Execute()
		if err == nil {
			t.Error("expected error for missing args")
		}
	})
}

func TestAssumptionsCommand(t *testing.T) {
	t.Run("generates assumptions prompt", func(t *testing.T) {
		prompt := generateAssumptionsPrompt("# Spec\n\n- Feature A", "test.md")

		if !strings.Contains(prompt, "implicit assumptions") {
			t.Error("expected prompt to contain 'implicit assumptions'")
		}
		if !strings.Contains(prompt, "Why It Matters") {
			t.Error("expected prompt to contain 'Why It Matters'")
		}
		if !strings.Contains(prompt, "What Breaks If It's Wrong") {
			t.Error("expected prompt to contain 'What Breaks If It's Wrong'")
		}
		if !strings.Contains(prompt, "How to Validate It") {
			t.Error("expected prompt to contain 'How to Validate It'")
		}
	})
}

func TestReviewCommand(t *testing.T) {
	t.Run("generates security persona prompt", func(t *testing.T) {
		prompt, err := generateReviewPrompt("# Spec", "test.md", "security")
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(prompt, "security reviewer") {
			t.Error("expected prompt to contain 'security reviewer'")
		}
		if !strings.Contains(prompt, "Attack Surfaces") {
			t.Error("expected prompt to contain 'Attack Surfaces'")
		}
	})

	t.Run("generates performance persona prompt", func(t *testing.T) {
		prompt, err := generateReviewPrompt("# Spec", "test.md", "performance")
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(prompt, "performance reviewer") {
			t.Error("expected prompt to contain 'performance reviewer'")
		}
		if !strings.Contains(prompt, "Bottlenecks") {
			t.Error("expected prompt to contain 'Bottlenecks'")
		}
	})

	t.Run("generates maintainer persona prompt", func(t *testing.T) {
		prompt, err := generateReviewPrompt("# Spec", "test.md", "maintainer")
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(prompt, "maintainability reviewer") {
			t.Error("expected prompt to contain 'maintainability reviewer'")
		}
	})

	t.Run("rejects unknown persona", func(t *testing.T) {
		_, err := generateReviewPrompt("# Spec", "test.md", "unknown")
		if err == nil {
			t.Error("expected error for unknown persona")
		}
		if !strings.Contains(err.Error(), "unknown persona") {
			t.Errorf("expected 'unknown persona' in error, got: %v", err)
		}
	})
}

func TestFailureCommand(t *testing.T) {
	t.Run("generates failure modes prompt", func(t *testing.T) {
		prompt := generateFailurePrompt("# Spec", "test.md")

		if !strings.Contains(prompt, "failure modes") {
			t.Error("expected prompt to contain 'failure modes'")
		}
		if !strings.Contains(prompt, "Partial Failures") {
			t.Error("expected prompt to contain 'Partial Failures'")
		}
		if !strings.Contains(prompt, "Race Conditions") {
			t.Error("expected prompt to contain 'Race Conditions'")
		}
		if !strings.Contains(prompt, "Misuse Cases") {
			t.Error("expected prompt to contain 'Misuse Cases'")
		}
	})
}

func TestTestIntentCommand(t *testing.T) {
	t.Run("generates test intent prompt", func(t *testing.T) {
		prompt := generateTestIntentPrompt("# Spec", "test.md")

		if !strings.Contains(prompt, "test intent") {
			t.Error("expected prompt to contain 'test intent'")
		}
		if !strings.Contains(prompt, "Do NOT generate test code") {
			t.Error("expected prompt to contain 'Do NOT generate test code'")
		}
		if !strings.Contains(prompt, "Required Invariants") {
			t.Error("expected prompt to contain 'Required Invariants'")
		}
		if !strings.Contains(prompt, "Boundary Conditions") {
			t.Error("expected prompt to contain 'Boundary Conditions'")
		}
		if !strings.Contains(prompt, "Negative Cases") {
			t.Error("expected prompt to contain 'Negative Cases'")
		}
	})
}

func TestTraceCommand(t *testing.T) {
	t.Run("generates trace prompt", func(t *testing.T) {
		prompt := generateTracePrompt("# Spec", "test.md")

		if !strings.Contains(prompt, "traceability mapping") {
			t.Error("expected prompt to contain 'traceability mapping'")
		}
		if !strings.Contains(prompt, "Identify Code Modules Affected") {
			t.Error("expected prompt to contain 'Identify Code Modules Affected'")
		}
		if !strings.Contains(prompt, "Dead or Obsolete Code Risks") {
			t.Error("expected prompt to contain 'Dead or Obsolete Code Risks'")
		}
	})
}

func TestDistillCommand(t *testing.T) {
	t.Run("generates exec audience prompt", func(t *testing.T) {
		prompt, err := generateDistillPrompt("# Spec", "test.md", "exec")
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(prompt, "executive audience") {
			t.Error("expected prompt to contain 'executive audience'")
		}
		if !strings.Contains(prompt, "Business Impact") {
			t.Error("expected prompt to contain 'Business Impact'")
		}
	})

	t.Run("generates implementer audience prompt", func(t *testing.T) {
		prompt, err := generateDistillPrompt("# Spec", "test.md", "implementer")
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(prompt, "implementation-focused developer") {
			t.Error("expected prompt to contain 'implementation-focused developer'")
		}
	})

	t.Run("generates ai audience prompt", func(t *testing.T) {
		prompt, err := generateDistillPrompt("# Spec", "test.md", "ai")
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(prompt, "AI coding assistant") {
			t.Error("expected prompt to contain 'AI coding assistant'")
		}
		if !strings.Contains(prompt, "Invariants") {
			t.Error("expected prompt to contain 'Invariants'")
		}
	})

	t.Run("generates qa audience prompt", func(t *testing.T) {
		prompt, err := generateDistillPrompt("# Spec", "test.md", "qa")
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(prompt, "QA engineer") {
			t.Error("expected prompt to contain 'QA engineer'")
		}
	})

	t.Run("rejects unknown audience", func(t *testing.T) {
		_, err := generateDistillPrompt("# Spec", "test.md", "unknown")
		if err == nil {
			t.Error("expected error for unknown audience")
		}
		if !strings.Contains(err.Error(), "unknown audience") {
			t.Errorf("expected 'unknown audience' in error, got: %v", err)
		}
	})
}

func TestCalibrateCommand(t *testing.T) {
	t.Run("generates default mode prompt", func(t *testing.T) {
		prompt, err := generateCalibratePrompt("# Spec\n\n- Feature A", "test.md", "default")
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(prompt, "Epistemic Calibration") {
			t.Error("expected prompt to contain 'Epistemic Calibration'")
		}
		if !strings.Contains(prompt, "High-Confidence Claims") {
			t.Error("expected prompt to contain 'High-Confidence Claims'")
		}
		if !strings.Contains(prompt, "Assumptions (Unproven but Required)") {
			t.Error("expected prompt to contain 'Assumptions (Unproven but Required)'")
		}
		if !strings.Contains(prompt, "Red Flags") {
			t.Error("expected prompt to contain 'Red Flags'")
		}
		if !strings.Contains(prompt, "Decision Checklist") {
			t.Error("expected prompt to contain 'Decision Checklist'")
		}
	})

	t.Run("generates confidence mode prompt", func(t *testing.T) {
		prompt, err := generateCalibratePrompt("# Spec\n\n- Feature A", "test.md", "confidence")
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(prompt, "Calibration Mode: Confidence") {
			t.Error("expected prompt to contain 'Calibration Mode: Confidence'")
		}
		if !strings.Contains(prompt, "High Confidence") {
			t.Error("expected prompt to contain 'High Confidence'")
		}
		if !strings.Contains(prompt, "Medium Confidence") {
			t.Error("expected prompt to contain 'Medium Confidence'")
		}
		if !strings.Contains(prompt, "Confidence Killers") {
			t.Error("expected prompt to contain 'Confidence Killers'")
		}
	})

	t.Run("generates uncertainty mode prompt", func(t *testing.T) {
		prompt, err := generateCalibratePrompt("# Spec", "test.md", "uncertainty")
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(prompt, "Calibration Mode: Uncertainty") {
			t.Error("expected prompt to contain 'Calibration Mode: Uncertainty'")
		}
		if !strings.Contains(prompt, "Uncertainty Register") {
			t.Error("expected prompt to contain 'Uncertainty Register'")
		}
		if !strings.Contains(prompt, "Ambiguity Hotspots") {
			t.Error("expected prompt to contain 'Ambiguity Hotspots'")
		}
	})

	t.Run("generates unknowns mode prompt", func(t *testing.T) {
		prompt, err := generateCalibratePrompt("# Spec", "test.md", "unknowns")
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(prompt, "Calibration Mode: Unknowns") {
			t.Error("expected prompt to contain 'Calibration Mode: Unknowns'")
		}
		if !strings.Contains(prompt, "Missing Information Inventory") {
			t.Error("expected prompt to contain 'Missing Information Inventory'")
		}
		if !strings.Contains(prompt, "Missing Decisions") {
			t.Error("expected prompt to contain 'Missing Decisions'")
		}
		if !strings.Contains(prompt, "Minimal Next-Step Questions") {
			t.Error("expected prompt to contain 'Minimal Next-Step Questions'")
		}
	})

	t.Run("rejects unknown mode", func(t *testing.T) {
		_, err := generateCalibratePrompt("# Spec", "test.md", "invalid")
		if err == nil {
			t.Error("expected error for unknown mode")
		}
		if !strings.Contains(err.Error(), "unknown mode") {
			t.Errorf("expected 'unknown mode' in error, got: %v", err)
		}
	})

	t.Run("includes artifact name and content", func(t *testing.T) {
		prompt, err := generateCalibratePrompt("Test content here", "/path/to/myartifact.md", "default")
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(prompt, "myartifact.md") {
			t.Error("expected prompt to contain artifact filename")
		}
		if !strings.Contains(prompt, "Test content here") {
			t.Error("expected prompt to contain artifact content")
		}
	})
}
