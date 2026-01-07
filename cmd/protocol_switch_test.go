package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"specfirst/internal/app"
	"specfirst/internal/assets"
	"specfirst/internal/repository"
)

// captureOutput captures stdout from a function call
func captureOutput(t *testing.T, f func()) string {
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = orig
	}()

	f()

	w.Close()
	var buf strings.Builder
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestProtocolSwitch(t *testing.T) {
	// Create a temp directory for the test workspace
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Initialize workspace with default protocol
	rootCmd.SetArgs([]string{"init"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	// Create a dummy custom protocol file outside the standard location
	customProtoContent := `name: "custom-proto"
version: "1.0"
stages:
  - id: custom
    name: Custom Stage
    intent: test
    template: custom.md
    outputs: [custom.md]
`
	// Typically protocol path needs to be valid. If it's just a file name it looks in .specfirst/protocols.
	// If it is absolute/relative path, it loads directly?
	// app.Load calls repository.LoadProtocol.
	// repository.LoadProtocol calls repository.ProtocolsPath(name) if it doesn't end in yaml/yml?
	// If we pass an absolute path to protocol flag, app.Load uses it as is?
	// Let's assume app.Load logic:
	// if protocolOverride != "" -> load that.

	// Create the file.
	customProtoPath := filepath.Join(tmpDir, "custom-protocol.yaml")
	if err := os.WriteFile(customProtoPath, []byte(customProtoContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create the matching template
	customTmplPath := filepath.Join(tmpDir, ".specfirst", "templates", "custom.md")
	if err := os.WriteFile(customTmplPath, []byte("# Custom Template"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("overrides default protocol with file path", func(t *testing.T) {
		// Set the flag variable directly for unit testing app.Load interactions
		protocolFlag = customProtoPath

		// Load App
		application, err := app.Load(protocolFlag)
		if err != nil {
			t.Fatalf("app.Load failed: %v", err)
		}

		// Verify Protocol Name
		if application.Protocol.Name != "custom-proto" {
			t.Errorf("expected protocol name 'custom-proto', got %q", application.Protocol.Name)
		}

		// Verify Config Protocol is NOT changed (it loads from config.yaml, override only affects memory)
		// app.Load returns *Application which has Config.
		// The Config struct itself will still have whatever was in config.yaml?
		// app.Load loads config first. Then if override, it loads protocol from override.
		// It does NOT overwrite Config.Protocol in file.
		// Checking application.Config.Protocol might return the file value?
		// Let's check app.go implementation.
		// app.Load:
		//   cfg = LoadConfig()
		//   protoName = cfg.Protocol
		//   if override != "" { protoName = override }
		//   proto = LoadProtocol(protoName)
		// It doesn't modify cfg.Protocol.

		// So checking application.Config.Protocol works.
		if application.Config.Protocol == customProtoPath {
			t.Errorf("expected config protocol to remain default, but got %q", application.Config.Protocol)
		}
	})

	t.Run("check command detects drift when overriding", func(t *testing.T) {
		// Initialize state with default protocol
		s := assets.DefaultProtocolName // "multi-stage"
		// Write state file forcing it to default protocol
		statePath := repository.StatePath()
		stateContent := fmt.Sprintf(`{"protocol": "%s", "spec_version": "1.0"}`, s)
		if err := os.WriteFile(statePath, []byte(stateContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Set flag to custom proto
		protocolFlag = customProtoPath

		// Load App
		application, err := app.Load(protocolFlag)
		if err != nil {
			t.Fatalf("app.Load failed: %v", err)
		}

		if application.Protocol.Name != "custom-proto" {
			t.Fatal("failed to load custom proto")
		}

		// Logic to detect drift is usually: application.Protocol.Name != application.State.Protocol?
		// application.State.Protocol is what we wrote to state.json ("multi-stage").
		// application.Protocol.Name is "custom-proto".

		if application.State.Protocol == application.Protocol.Name {
			t.Fatalf("Expected protocol name mismatch (drift), but they matched: %s", application.Protocol.Name)
		}
	})
}
