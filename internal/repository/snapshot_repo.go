package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"specfirst/internal/domain"
	"specfirst/internal/utils"
)

// Metadata defines the schema for snapshot metadata
type Metadata struct {
	Version         string    `json:"version"`
	Protocol        string    `json:"protocol"`
	ArchivedAt      time.Time `json:"archived_at"`
	StagesCompleted []string  `json:"stages_completed"`
	Tags            []string  `json:"tags,omitempty"`
	Notes           string    `json:"notes,omitempty"`
}

// CreateParams holds the pre-loaded dependencies for snapshot creation.
type CreateParams struct {
	Config   domain.Config
	Protocol domain.Protocol
	State    domain.State
}

// SnapshotRepository handles snapshot operations (archives or tracks)
type SnapshotRepository struct {
	RootDir string
}

func NewSnapshotRepository(rootDir string) *SnapshotRepository {
	return &SnapshotRepository{RootDir: rootDir}
}

var validVersionPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)

// Removed local ValidateSnapshotName as it corresponds to domain logic.

func (r *SnapshotRepository) List() ([]string, error) {
	entries, err := os.ReadDir(r.RootDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	versions := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			versions = append(versions, entry.Name())
		}
	}
	sort.Strings(versions)
	return versions, nil
}

func (r *SnapshotRepository) Create(version string, tags []string, notes string, params CreateParams) error {
	if !domain.IsValidSnapshotName(version) {
		return fmt.Errorf("invalid snapshot name: %s", version)
	}

	proto := params.Protocol
	s := params.State

	// Validate artifacts exist
	for stageID, output := range s.StageOutputs {
		for _, artifactPath := range output.Files {
			abs, err := ArtifactAbsFromState(artifactPath)
			if err != nil {
				return fmt.Errorf("invalid artifact path %s: %w", artifactPath, err)
			}
			if _, err := os.Stat(abs); err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("missing artifact for stage %s: %s (snapshot aborted)", stageID, artifactPath)
				}
				return fmt.Errorf("cannot access artifact for stage %s: %s (%v)", stageID, artifactPath, err)
			}
		}
	}

	snapshotRoot := filepath.Join(r.RootDir, version)
	tmpRoot := snapshotRoot + ".tmp"

	if err := utils.EnsureDir(filepath.Dir(snapshotRoot)); err != nil {
		return err
	}
	_ = os.RemoveAll(tmpRoot)

	if _, err := os.Stat(snapshotRoot); err == nil {
		return fmt.Errorf("snapshot already exists: %s", version)
	}
	if err := os.Mkdir(tmpRoot, 0755); err != nil {
		return err
	}

	cleanup := true
	defer func() {
		if cleanup {
			_ = os.RemoveAll(tmpRoot)
		}
	}()

	if err := utils.CopyDir(ArtifactsPath(), filepath.Join(tmpRoot, "artifacts")); err != nil {
		return err
	}
	if err := utils.CopyDir(GeneratedPath(), filepath.Join(tmpRoot, "generated")); err != nil {
		return err
	}
	if err := utils.CopyDirWithOpts(ProtocolsPath(), filepath.Join(tmpRoot, "protocols"), true); err != nil {
		return err
	}
	if err := utils.CopyDirWithOpts(TemplatesPath(), filepath.Join(tmpRoot, "templates"), true); err != nil {
		return err
	}
	if err := utils.CopyFile(ConfigPath(), filepath.Join(tmpRoot, "config.yaml")); err != nil {
		return err
	}
	if err := utils.CopyFile(StatePath(), filepath.Join(tmpRoot, "state.json")); err != nil {
		return err
	}

	metadata := Metadata{
		Version:         version,
		Protocol:        proto.Name,
		ArchivedAt:      time.Now().UTC(),
		StagesCompleted: s.CompletedStages,
		Tags:            tags,
		Notes:           notes,
	}
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(filepath.Join(tmpRoot, "metadata.json"), data, 0644); err != nil {
		return err
	}

	if err := os.Rename(tmpRoot, snapshotRoot); err != nil {
		return err
	}

	cleanup = false
	return nil
}

func (r *SnapshotRepository) Restore(version string, force bool) error {
	if !domain.IsValidSnapshotName(version) {
		return fmt.Errorf("invalid snapshot name: %s", version)
	}
	snapshotRoot := filepath.Join(r.RootDir, version)

	if info, err := os.Stat(snapshotRoot); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("snapshot not found: %s", version)
		}
		return err
	} else if !info.IsDir() {
		return fmt.Errorf("snapshot is not a directory: %s", version)
	}

	existingPaths := []string{
		ArtifactsPath(),
		GeneratedPath(),
		ProtocolsPath(),
		TemplatesPath(),
		ConfigPath(),
		StatePath(),
	}

	// Staging
	restoreStaging := SpecPath() + "_restore.tmp"
	_ = os.RemoveAll(restoreStaging)
	if err := utils.EnsureDir(restoreStaging); err != nil {
		return fmt.Errorf("failed to create restore staging directory: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(restoreStaging)
	}()

	// Stage components
	if err := utils.CopyDir(filepath.Join(snapshotRoot, "artifacts"), filepath.Join(restoreStaging, "artifacts")); err != nil {
		return fmt.Errorf("failed to stage artifacts: %w", err)
	}
	if err := utils.CopyDir(filepath.Join(snapshotRoot, "generated"), filepath.Join(restoreStaging, "generated")); err != nil {
		return fmt.Errorf("failed to stage generated: %w", err)
	}
	if err := utils.CopyDirWithOpts(filepath.Join(snapshotRoot, "protocols"), filepath.Join(restoreStaging, "protocols"), true); err != nil {
		return fmt.Errorf("failed to stage protocols: %w", err)
	}
	if err := utils.CopyDirWithOpts(filepath.Join(snapshotRoot, "templates"), filepath.Join(restoreStaging, "templates"), true); err != nil {
		return fmt.Errorf("failed to stage templates: %w", err)
	}
	if err := utils.CopyFile(filepath.Join(snapshotRoot, "config.yaml"), filepath.Join(restoreStaging, "config.yaml")); err != nil {
		return fmt.Errorf("failed to stage config: %w", err)
	}
	if err := utils.CopyFile(filepath.Join(snapshotRoot, "state.json"), filepath.Join(restoreStaging, "state.json")); err != nil {
		return fmt.Errorf("failed to stage state: %w", err)
	}

	archivedConfigPath := filepath.Join(snapshotRoot, "config.yaml")
	metadataPath := filepath.Join(snapshotRoot, "metadata.json")
	metadataData, err := os.ReadFile(metadataPath)
	if err != nil {
		return fmt.Errorf("cannot read archive metadata: %w", err)
	}
	var metadata Metadata
	if err := json.Unmarshal(metadataData, &metadata); err != nil {
		return fmt.Errorf("cannot parse archive metadata: %w", err)
	}

	archivedCfg, err := LoadConfig(archivedConfigPath)
	if err != nil {
		return fmt.Errorf("cannot load archived config: %w", err)
	}
	if strings.TrimSpace(archivedCfg.Protocol) == "" {
		return fmt.Errorf("archive is incomplete or corrupt: config missing protocol")
	}

	archivedProtoPath := filepath.Join(snapshotRoot, "protocols", archivedCfg.Protocol+".yaml")
	if _, err := os.Stat(archivedProtoPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("archive is incomplete or corrupt: missing protocol file %s", filepath.Base(archivedProtoPath))
		}
		return fmt.Errorf("cannot access archived protocol file %s: %w", filepath.Base(archivedProtoPath), err)
	}
	archivedProto, err := LoadProtocol(archivedProtoPath)
	if err != nil {
		return fmt.Errorf("cannot load archived protocol: %w", err)
	}
	if metadata.Protocol != "" && archivedProto.Name != metadata.Protocol {
		return fmt.Errorf("archive metadata protocol mismatch: metadata=%s protocol=%s", metadata.Protocol, archivedProto.Name)
	}

	// Swap logic
	type backupEntry struct {
		original string
		backup   string
	}
	var backups []backupEntry
	var createdPaths []string // Track created paths for cleaner rollback

	success := false
	defer func() {
		if !success {
			// Rollback
			// 1. Restore backups
			for i := len(backups) - 1; i >= 0; i-- {
				entry := backups[i]
				_ = os.RemoveAll(entry.original)
				_ = os.Rename(entry.backup, entry.original)
			}
			// 2. Cleanup created paths (files/dirs that didn't exist before and thus have no backup)
			for _, path := range createdPaths {
				_ = os.RemoveAll(path)
			}
		} else {
			// Clean backups
			for _, entry := range backups {
				_ = os.RemoveAll(entry.backup)
			}
		}
	}()

	for _, path := range existingPaths {
		stagedPath := filepath.Join(restoreStaging, filepath.Base(path))

		// If existing component exists, backup
		if _, err := os.Stat(path); err == nil {
			if !force {
				return fmt.Errorf("workspace has data at %s; use --force to overwrite", path)
			}
			oldPath := path + ".old"
			_ = os.RemoveAll(oldPath)
			if err := os.Rename(path, oldPath); err != nil {
				return fmt.Errorf("failed to backup existing %s: %w", filepath.Base(path), err)
			}
			backups = append(backups, backupEntry{original: path, backup: oldPath})
		} else {
			// Path didn't exist, we will create it (by moving staged).
			// If we fail later, we must remove it.
			createdPaths = append(createdPaths, path)
		}

		// Move staged to path
		if _, err := os.Stat(stagedPath); err == nil {
			if err := os.Rename(stagedPath, path); err != nil {
				return fmt.Errorf("failed to restore %s: %w", filepath.Base(path), err)
			}
		}
	}

	success = true
	return nil
}

func (r *SnapshotRepository) Compare(leftVersion, rightVersion string) ([]string, []string, []string, error) {
	if !domain.IsValidSnapshotName(leftVersion) {
		return nil, nil, nil, fmt.Errorf("invalid snapshot name: %s", leftVersion)
	}
	if !domain.IsValidSnapshotName(rightVersion) {
		return nil, nil, nil, fmt.Errorf("invalid snapshot name: %s", rightVersion)
	}

	leftRoot := filepath.Join(r.RootDir, leftVersion, "artifacts")
	rightRoot := filepath.Join(r.RootDir, rightVersion, "artifacts")

	left, err := utils.CollectFileHashes(leftRoot)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to collect hashes for %s: %w", leftVersion, err)
	}
	right, err := utils.CollectFileHashes(rightRoot)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to collect hashes for %s: %w", rightVersion, err)
	}

	added, removed, changed := compareHashes(left, right)
	return added, removed, changed, nil
}

func compareHashes(left map[string]string, right map[string]string) ([]string, []string, []string) {
	added := []string{}
	removed := []string{}
	changed := []string{}

	for path, hash := range right {
		if leftHash, ok := left[path]; ok {
			if leftHash != hash {
				changed = append(changed, path)
			}
		} else {
			added = append(added, path)
		}
	}
	for path := range left {
		if _, ok := right[path]; !ok {
			removed = append(removed, path)
		}
	}

	sort.Strings(added)
	sort.Strings(removed)
	sort.Strings(changed)
	return added, removed, changed
}
