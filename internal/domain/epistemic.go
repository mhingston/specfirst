package domain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

func generateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *State) AddAssumption(text, owner string) string {
	id := generateID()
	a := Assumption{
		ID:        id,
		Text:      text,
		Status:    "open",
		Owner:     owner,
		CreatedAt: time.Now(),
	}
	s.Epistemics.Assumptions = append(s.Epistemics.Assumptions, a)
	return id
}

func (s *State) CloseAssumption(id, status string) bool {
	for i := range s.Epistemics.Assumptions {
		if s.Epistemics.Assumptions[i].ID == id {
			s.Epistemics.Assumptions[i].Status = status
			return true
		}
	}
	return false
}

func (s *State) AddOpenQuestion(text string, tags []string, context string) string {
	id := generateID()
	q := OpenQuestion{
		ID:      id,
		Text:    text,
		Tags:    tags,
		Status:  "open",
		Context: context,
	}
	s.Epistemics.OpenQuestions = append(s.Epistemics.OpenQuestions, q)
	return id
}

func (s *State) ResolveOpenQuestion(id, answer string) bool {
	for i := range s.Epistemics.OpenQuestions {
		if s.Epistemics.OpenQuestions[i].ID == id {
			s.Epistemics.OpenQuestions[i].Status = "resolved"
			s.Epistemics.OpenQuestions[i].Answer = answer
			return true
		}
	}
	return false
}

func (s *State) AddDecision(text, rationale string, alternatives []string) string {
	id := generateID()
	d := Decision{
		ID:           id,
		Text:         text,
		Rationale:    rationale,
		Alternatives: alternatives,
		Status:       "accepted", // default?
		CreatedAt:    time.Now(),
	}
	s.Epistemics.Decisions = append(s.Epistemics.Decisions, d)
	return id
}

func (s *State) UpdateDecision(id, status string) bool {
	for i := range s.Epistemics.Decisions {
		if s.Epistemics.Decisions[i].ID == id {
			s.Epistemics.Decisions[i].Status = status
			return true
		}
	}
	return false
}

func (s *State) AddRisk(text, severity string) string {
	id := generateID()
	r := Risk{
		ID:       id,
		Text:     text,
		Severity: severity,
		Status:   "open",
	}
	s.Epistemics.Risks = append(s.Epistemics.Risks, r)
	return id
}

func (s *State) MitigateRisk(id, mitigation, status string) bool {
	for i := range s.Epistemics.Risks {
		if s.Epistemics.Risks[i].ID == id {
			s.Epistemics.Risks[i].Mitigation = mitigation
			s.Epistemics.Risks[i].Status = status
			return true
		}
	}
	return false
}

func (s *State) AddDispute(topic string) string {
	id := generateID()
	d := Dispute{
		ID:     id,
		Topic:  topic,
		Status: "open",
	}
	s.Epistemics.Disputes = append(s.Epistemics.Disputes, d)
	return id
}

func (s *State) ResolveDispute(id string) bool {
	for i := range s.Epistemics.Disputes {
		if s.Epistemics.Disputes[i].ID == id {
			s.Epistemics.Disputes[i].Status = "resolved"
			return true
		}
	}
	return false
}

// MissingApprovals checks which required approvals are missing from state
func MissingApprovals(required []Approval, s State) []string {
	missing := []string{}
	for _, req := range required {
		if s.IsStageCompleted(req.Stage) {
			if !s.HasAttestation(req.Stage, req.Role, "approved") {
				missing = append(missing, fmt.Sprintf("%s (role: %s)", req.Stage, req.Role))
			}
		}
	}
	return missing
}
