package filter

import (
	"time"

	"github.com/GustavoSchmitzz/vagas-radar/internal/config"
	"github.com/GustavoSchmitzz/vagas-radar/internal/domain"
)

// Engine coordena a aplicação do pipeline de regras de negócio.
type Engine struct {
	internshipRule *InternshipRule
	locationRule   *LocationRule
}

// NewEngine inicializa a engine com as configurações fornecidas.
func NewEngine(cfg config.FilterConfig) *Engine {
	return &Engine{
		internshipRule: NewInternshipRule(cfg),
		locationRule:   NewLocationRule(cfg),
	}
}

// EvaluationResult agrega a decisão final e os dados normalizados da vaga.
type EvaluationResult struct {
	Job      *domain.Job
	Accepted bool
	Reason   string
}

// Evaluate submete uma vaga bruta ao pipeline de regras:
// 1. Validação de estágio / processos seletivos e senioridades permitidas.
// 2. Validação de modalidade e critérios geográficos estritos.
func (e *Engine) Evaluate(raw domain.RawJob) EvaluationResult {
	// Passo 1: Avaliar senioridade, tipo de oportunidade e contrato
	classRes := e.internshipRule.Classify(raw)
	if !classRes.Accepted {
		return EvaluationResult{
			Job:      nil,
			Accepted: false,
			Reason:   classRes.Reason,
		}
	}

	// Passo 2: Avaliar modalidade e restrições de localidade
	locRes := e.locationRule.Evaluate(raw)
	if !locRes.Accepted {
		return EvaluationResult{
			Job:      nil,
			Accepted: false,
			Reason:   locRes.Reason,
		}
	}

	// Publicação da vaga
	pubDate := raw.PublishedAt
	if pubDate.IsZero() {
		pubDate = time.Now()
	}

	// Construir a entidade final
	job := &domain.Job{
		Title:           raw.Title,
		Company:         raw.Company,
		Description:     raw.Description,
		URL:             raw.URL,
		Source:          raw.Source,
		Location:        locRes.Location,
		WorkMode:        locRes.WorkMode,
		OpportunityType: classRes.OpportunityType,
		Seniority:       classRes.Seniority,
		ContractType:    classRes.ContractType,
		PublishedAt:     pubDate,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Calcula o hash de deduplicação único
	job.ComputeHash()

	return EvaluationResult{
		Job:      job,
		Accepted: true,
		Reason:   locRes.Reason,
	}
}
