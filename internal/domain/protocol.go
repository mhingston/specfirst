package domain

// Protocol represents a workflow definition with stages and approvals.
type Protocol struct {
	Name      string      `yaml:"name"`
	Version   string      `yaml:"version"`
	Uses      []string    `yaml:"uses,omitempty"` // Protocol imports/mixins
	Stages    []Stage     `yaml:"stages"`
	Approvals []Approval  `yaml:"approvals"`
	Lint      *LintConfig `yaml:"lint,omitempty"` // Protocol-level schema additions
}

// Stage represents a workflow step with optional type, modifiers, and contracts.
type Stage struct {
	ID        string   `yaml:"id"`
	Name      string   `yaml:"name"`
	Type      string   `yaml:"type,omitempty"` // "spec", "decompose", "task_prompt"
	Intent    string   `yaml:"intent"`
	Template  string   `yaml:"template"`
	DependsOn []string `yaml:"depends_on"`
	Inputs    []string `yaml:"inputs"`
	Outputs   []string `yaml:"outputs"`

	// Stage modifiers
	Optional   bool `yaml:"optional,omitempty"`
	Repeatable bool `yaml:"repeatable,omitempty"`
	Terminal   bool `yaml:"terminal,omitempty"`

	// Prompt configuration
	Prompt *PromptConfig `yaml:"prompt,omitempty"`

	// Output contract
	Output *OutputContract `yaml:"output,omitempty"`

	// For task_prompt type - reference to decompose stage
	Source string `yaml:"source,omitempty"`

	// Ambiguity Gates
	MaxOpenQuestions        *int     `yaml:"max_open_questions,omitempty"`
	MustResolveTags         []string `yaml:"must_resolve_tags,omitempty"`
	MaxHighRisksUnmitigated *int     `yaml:"max_high_risks_unmitigated,omitempty"`
}

// PromptConfig defines prompt generation parameters.
type PromptConfig struct {
	Intent            string      `yaml:"intent,omitempty"`
	ExpectedOutput    string      `yaml:"expected_output,omitempty"`
	Determinism       string      `yaml:"determinism,omitempty"`        // high, medium, low
	AllowedCreativity string      `yaml:"allowed_creativity,omitempty"` // high, medium, low
	Granularity       string      `yaml:"granularity,omitempty"`        // feature, story, ticket, commit
	MaxTasks          int         `yaml:"max_tasks,omitempty"`
	PreferParallel    bool        `yaml:"prefer_parallel,omitempty"`
	RiskBias          string      `yaml:"risk_bias,omitempty"` // conservative, balanced, fast
	Rules             []string    `yaml:"rules,omitempty"`
	RequiredFields    []string    `yaml:"required_fields,omitempty"`
	Lint              *LintConfig `yaml:"lint,omitempty"` // Stage-level schema additions
}

// LintConfig defines additional validation rules for prompts.
type LintConfig struct {
	RequiredSections []string `yaml:"required_sections,omitempty"`
	ForbiddenPhrases []string `yaml:"forbidden_phrases,omitempty"`
}

// OutputContract defines expected output structure.
type OutputContract struct {
	Format         string   `yaml:"format,omitempty"` // markdown, yaml, json
	Sections       []string `yaml:"sections,omitempty"`
	RequiredFields []string `yaml:"required_fields,omitempty"`
}

type Approval struct {
	Role  string `yaml:"role"`
	Stage string `yaml:"stage"`
}

func (p Protocol) StageByID(id string) (Stage, bool) {
	for _, stage := range p.Stages {
		if stage.ID == id {
			return stage, true
		}
	}
	return Stage{}, false
}

// NextStage returns the next stage in the sequence after the given stage ID.
func (p Protocol) NextStage(currentID string) *Stage {
	for i, stage := range p.Stages {
		if stage.ID == currentID {
			if i+1 < len(p.Stages) {
				return &p.Stages[i+1]
			}
			return nil
		}
	}
	// If currentID not found, maybe return first stage? Or nil.
	// Logic in CompleteStage was: "if currentID matches, check next".
	return nil
}

// ValidSnapshotNamePattern is the regex pattern for snapshot names.
const ValidSnapshotNamePattern = `^[a-zA-Z0-9][a-zA-Z0-9._-]*$`

// IsValidSnapshotName checks if a snapshot name is valid.
func IsValidSnapshotName(name string) bool {
	if name == "" {
		return false
	}
	if len(name) > 128 {
		return false
	}
	// Check first char
	first := name[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || (first >= '0' && first <= '9')) {
		return false
	}
	// Check rest
	for i := 1; i < len(name); i++ {
		c := name[i]
		isAlphaNum := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
		if !isAlphaNum && c != '.' && c != '_' && c != '-' {
			return false
		}
	}
	// Check for traversal (though . and .. are caught by structure generally, but explicit check doesn't hurt)
	// Actually .. is a dot then dot.
	for i := 0; i < len(name)-1; i++ {
		if name[i] == '.' && name[i+1] == '.' {
			return false
		}
	}
	return true
}
