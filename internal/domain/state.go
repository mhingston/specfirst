package domain

import (
	"time"
)

type State struct {
	Protocol        string                   `json:"protocol"`
	CurrentStage    string                   `json:"current_stage"`
	CompletedStages []string                 `json:"completed_stages"`
	StartedAt       time.Time                `json:"started_at"`
	SpecVersion     string                   `json:"spec_version"`
	StageOutputs    map[string]StageOutput   `json:"stage_outputs"`
	Attestations    map[string][]Attestation `json:"attestations"`
	Epistemics      Epistemics               `json:"epistemics,omitempty"`
}

type Epistemics struct {
	Assumptions   []Assumption   `json:"assumptions,omitempty"`
	OpenQuestions []OpenQuestion `json:"open_questions,omitempty"`
	Decisions     []Decision     `json:"decisions,omitempty"`
	Risks         []Risk         `json:"risks,omitempty"`
	Disputes      []Dispute      `json:"disputes,omitempty"`
	Confidence    Confidence     `json:"confidence,omitempty"`
}

type Assumption struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	Status    string    `json:"status"` // open, validated, invalidated
	Owner     string    `json:"owner,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type OpenQuestion struct {
	ID      string   `json:"id"`
	Text    string   `json:"text"`
	Tags    []string `json:"tags,omitempty"`
	Status  string   `json:"status"` // open, resolved, deferred
	Answer  string   `json:"answer,omitempty"`
	Context string   `json:"context,omitempty"` // file or section reference
}

type Decision struct {
	ID           string    `json:"id"`
	Text         string    `json:"text"`
	Rationale    string    `json:"rationale"`
	Alternatives []string  `json:"alternatives,omitempty"`
	Status       string    `json:"status"` // proposed, accepted, reversed
	CreatedAt    time.Time `json:"created_at"`
}

type Risk struct {
	ID         string `json:"id"`
	Text       string `json:"text"`
	Severity   string `json:"severity"` // low, medium, high
	Mitigation string `json:"mitigation,omitempty"`
	Status     string `json:"status"` // open, mitigated, accepted
}

type Dispute struct {
	ID        string     `json:"id"`
	Topic     string     `json:"topic"`
	Positions []Position `json:"positions,omitempty"`
	Status    string     `json:"status"` // open, resolved
}

type Position struct {
	Owner string `json:"owner"`
	Claim string `json:"claim"`
}

type Confidence struct {
	Overall string            `json:"overall"` // low, medium, high
	ByStage map[string]string `json:"by_stage,omitempty"`
}

type StageOutput struct {
	CompletedAt time.Time `json:"completed_at"`
	Files       []string  `json:"files"`
	PromptHash  string    `json:"prompt_hash"`
}

type Attestation struct {
	Role       string    `json:"role"`
	AttestedBy string    `json:"attested_by"`
	Scope      []string  `json:"scope"`
	Status     string    `json:"status"` // approved, approved_with_conditions, needs_changes, rejected
	Conditions []string  `json:"conditions,omitempty"`
	Rationale  string    `json:"rationale"`
	Date       time.Time `json:"date"`
}

func NewState(protocol string) State {
	return State{
		Protocol:        protocol,
		StartedAt:       time.Now(),
		CompletedStages: []string{},
		StageOutputs:    make(map[string]StageOutput),
		Attestations:    make(map[string][]Attestation),
		Epistemics: Epistemics{
			Assumptions:   []Assumption{},
			OpenQuestions: []OpenQuestion{},
			Decisions:     []Decision{},
			Risks:         []Risk{},
			Disputes:      []Dispute{},
			Confidence: Confidence{
				ByStage: make(map[string]string),
			},
		},
	}
}

func (s State) IsStageCompleted(id string) bool {
	for _, stage := range s.CompletedStages {
		if stage == id {
			return true
		}
	}
	return false
}

func (s *State) AddAttestation(stageID string, attestation Attestation) {
	if s.Attestations == nil {
		s.Attestations = make(map[string][]Attestation)
	}
	s.Attestations[stageID] = append(s.Attestations[stageID], attestation)
}

func (s State) HasAttestation(stageID, role, status string) bool {
	if s.Attestations == nil {
		return false
	}
	attestations, ok := s.Attestations[stageID]
	if !ok {
		return false
	}
	for _, a := range attestations {
		if a.Role == role && a.Status == status {
			return true
		}
	}
	return false
}
