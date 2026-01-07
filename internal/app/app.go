package app

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"specfirst/internal/assets"
	"specfirst/internal/domain"
	"specfirst/internal/engine/system"
	"specfirst/internal/repository"
)

// Application is the main application coordinator.
type Application struct {
	Config   domain.Config
	Protocol domain.Protocol
	State    domain.State
}

// NewApplication creates a new Application instance.
func NewApplication(cfg domain.Config, proto domain.Protocol, s domain.State) *Application {
	return &Application{
		Config:   cfg,
		Protocol: proto,
		State:    s,
	}
}

// Load loads the application state from the filesystem.
func Load(protocolOverride string) (*Application, error) {
	// Initialize System Assets (Templates)
	if err := system.Load(); err != nil {
		return nil, fmt.Errorf("loading system assets: %w", err)
	}

	// 1. Load Config
	cfg, err := repository.LoadConfig(repository.ConfigPath())
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	// 2. Resolve Active Protocol
	activeProto := resolveActiveProtocol(cfg, protocolOverride)
	if activeProto == "" {
		activeProto = assets.DefaultProtocolName
	}

	// 3. Load Protocol
	var protoPath string
	if filepath.IsAbs(activeProto) || strings.HasPrefix(activeProto, "./") || strings.HasPrefix(activeProto, "../") {
		protoPath = activeProto
	} else {
		protoPath = repository.ProtocolsPath(activeProto + ".yaml")
	}

	proto, err := repository.LoadProtocol(protoPath)
	if err != nil {
		return nil, fmt.Errorf("loading protocol %s: %w", activeProto, err)
	}

	// 4. Load State
	s, err := repository.LoadState(repository.StatePath())
	if err != nil {
		return nil, fmt.Errorf("loading state: %w", err)
	}

	// 5. Initialize State if needed
	if s.Protocol == "" {
		s = domain.NewState(proto.Name)
	} else if s.Protocol != proto.Name {
		// Just warn? Or update?
		// For now we assume state protocol is the source of truth for what started.
		// If config changed, we might have a mismatch.
	}
	// Initializing CurrentStage if empty
	if s.CurrentStage == "" && len(proto.Stages) > 0 {
		s.CurrentStage = proto.Stages[0].ID
	}

	return NewApplication(cfg, proto, s), nil
}

func resolveActiveProtocol(cfg domain.Config, override string) string {
	if override != "" {
		return override
	}
	if cfg.Protocol != "" {
		return cfg.Protocol
	}
	return ""
}

// AttestStage records an attestation for a stage.
func (app *Application) AttestStage(stageID, role, user, status, notes string, conditions []string) ([]string, error) {
	var warnings []string

	// 1. Verify approval is declared in protocol
	declared := false
	for _, approval := range app.Protocol.Approvals {
		if approval.Stage == stageID && approval.Role == role {
			declared = true
			break
		}
	}
	if !declared {
		return nil, fmt.Errorf("approval not declared in protocol: stage=%s role=%s", stageID, role)
	}

	// 2. Warn if stage is not completed (but allow it)
	if !app.State.IsStageCompleted(stageID) {
		warnings = append(warnings, fmt.Sprintf("stage %s is not yet completed; attestation recorded preemptively", stageID))
	}

	// 3. Record approval in state (as Attestation)
	attestation := domain.Attestation{
		Role:       role,
		AttestedBy: user,
		Status:     status,
		Rationale:  notes,
		Conditions: conditions,
		Date:       time.Now().UTC(),
	}
	app.State.AddAttestation(stageID, attestation)

	// 4. Save state
	return warnings, app.SaveState()
}

// SaveState saves the current state to disk.
func (app *Application) SaveState() error {
	return repository.SaveState(repository.StatePath(), app.State)
}
