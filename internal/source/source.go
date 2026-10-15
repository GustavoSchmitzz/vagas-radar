package source

import (
	"context"

	"github.com/GustavoSchmitzz/vagas-radar/internal/domain"
)

// JobSource é o contrato obrigatório que qualquer provedor de vagas ou processos seletivos deve implementar.
type JobSource interface {
	// Name retorna o identificador legível da fonte (ex: "github:backend-br", "feed:programas").
	Name() string

	// Fetch consulta a fonte externa e retorna a listagem de vagas brutas coletadas.
	Fetch(ctx context.Context) ([]domain.RawJob, error)
}

// Registry gerencia a coleção de fontes registradas na aplicação.
type Registry struct {
	sources []JobSource
}

// NewRegistry inicializa um registro vazio de fontes.
func NewRegistry() *Registry {
	return &Registry{
		sources: make([]JobSource, 0),
	}
}

// Register adiciona uma nova fonte ao catálogo de varredura.
func (r *Registry) Register(s JobSource) {
	r.sources = append(r.sources, s)
}

// All retorna todas as fontes registradas.
func (r *Registry) All() []JobSource {
	return r.sources
}
