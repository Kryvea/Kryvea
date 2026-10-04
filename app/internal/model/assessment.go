package model

import "time"

const (
	AssessmentStatusOnHold     = "On Hold"
	AssessmentStatusInProgress = "In Progress"
	AssessmentStatusCompleted  = "Completed"
)

type Assessment struct {
	Model
	Name               string          `json:"name,omitempty"`
	Language           string          `json:"language,omitempty"`
	StartDateTime      time.Time       `json:"start_date_time,omitempty"`
	EndDateTime        time.Time       `json:"end_date_time,omitempty"`
	KickoffDateTime    time.Time       `json:"kickoff_date_time,omitempty"`
	Targets            []Target        `json:"targets"`
	Status             string          `json:"status,omitempty"`
	Type               AssessmentType  `json:"type,omitempty"`
	CVSSVersions       map[string]bool `json:"cvss_versions,omitempty"`
	Environment        string          `json:"environment,omitempty"`
	TestingType        string          `json:"testing_type,omitempty"`
	OSSTMMVector       string          `json:"osstmm_vector,omitempty"`
	VulnerabilityCount int             `json:"vulnerability_count,omitempty"`
	Customer           Customer        `json:"customer,omitempty"`
	IsOwned            bool            `json:"is_owned,omitempty"`
}

type AssessmentType struct {
	Short string `json:"short"`
	Full  string `json:"full"`
}
