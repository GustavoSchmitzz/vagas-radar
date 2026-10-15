package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// WorkMode define a modalidade de trabalho da vaga.
type WorkMode string

const (
	WorkModeRemote  WorkMode = "REMOTE"
	WorkModeOnSite  WorkMode = "ON_SITE"
	WorkModeHybrid  WorkMode = "HYBRID"
	WorkModeUnknown WorkMode = "UNKNOWN"
)

// OpportunityType categoriza o tipo de oportunidade encontrada.
type OpportunityType string

const (
	// OpportunityInternship representa uma vaga de estágio regular.
	OpportunityInternship OpportunityType = "INTERNSHIP"
	// OpportunitySelectionProcess representa programas corporativos ou processos seletivos formais de estágio.
	OpportunitySelectionProcess OpportunityType = "SELECTION_PROCESS"
	// OpportunityGeneralJob representa uma vaga de trabalho comum (não-estágio).
	OpportunityGeneralJob OpportunityType = "GENERAL_JOB"
)

// Seniority define o nível de experiência exigido.
type Seniority string

const (
	SeniorityInternship Seniority = "INTERNSHIP"
	SeniorityJunior     Seniority = "JUNIOR"
	SeniorityMid        Seniority = "MID"
	SenioritySenior     Seniority = "SENIOR"
	SeniorityUnknown    Seniority = "UNKNOWN"
)

// ContractType define o regime de contratação.
type ContractType string

const (
	ContractInternship ContractType = "INTERNSHIP"
	ContractCLT        ContractType = "CLT"
	ContractPJ         ContractType = "PJ"
	ContractUnknown    ContractType = "UNKNOWN"
)

// Location agrupa os detalhes geográficos e restrições da oportunidade.
type Location struct {
	City             string   `json:"city"`
	State            string   `json:"state"`
	Country          string   `json:"country"`
	Raw              string   `json:"raw"`
	RestrictedStates []string `json:"restricted_states,omitempty"`
}

// RawJob representa o payload bruto obtido diretamente de uma fonte externa (JobSource).
type RawJob struct {
	ExternalID  string            `json:"external_id"`
	Title       string            `json:"title"`
	Company     string            `json:"company"`
	Description string            `json:"description"`
	URL         string            `json:"url"`
	LocationRaw string            `json:"location_raw"`
	WorkModeRaw string            `json:"work_mode_raw"`
	Source      string            `json:"source"`
	PublishedAt time.Time         `json:"published_at"`
	Extra       map[string]string `json:"extra,omitempty"`
}

// Job representa uma oportunidade validada, normalizada e persistível.
type Job struct {
	ID              int64           `json:"id"`
	Hash            string          `json:"hash"`
	Title           string          `json:"title"`
	Company         string          `json:"company"`
	Description     string          `json:"description"`
	URL             string          `json:"url"`
	Source          string          `json:"source"`
	Location        Location        `json:"location"`
	WorkMode        WorkMode        `json:"work_mode"`
	OpportunityType OpportunityType `json:"opportunity_type"`
	Seniority       Seniority       `json:"seniority"`
	ContractType    ContractType    `json:"contract_type"`
	PublishedAt     time.Time       `json:"published_at"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// GenerateHash gera um hash SHA-256 determinístico para deduplicação da oportunidade.
// Se houver URL válida, usa a URL canônica normalizada. Caso contrário, combina empresa e título.
func GenerateHash(source, company, title, rawURL string) string {
	cleanURL := strings.TrimSpace(strings.ToLower(rawURL))
	cleanCompany := strings.TrimSpace(strings.ToLower(company))
	cleanTitle := strings.TrimSpace(strings.ToLower(title))
	cleanSource := strings.TrimSpace(strings.ToLower(source))

	var seed string
	if cleanURL != "" {
		seed = fmt.Sprintf("%s|%s", cleanSource, cleanURL)
	} else {
		seed = fmt.Sprintf("%s|%s|%s", cleanSource, cleanCompany, cleanTitle)
	}

	hash := sha256.Sum256([]byte(seed))
	return hex.EncodeToString(hash[:])
}

// ComputeHash calcula e preenche o hash da instância de Job.
func (j *Job) ComputeHash() {
	j.Hash = GenerateHash(j.Source, j.Company, j.Title, j.URL)
}
