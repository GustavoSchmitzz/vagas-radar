package storage

import (
	"context"

	"github.com/GustavoSchmitzz/vagas-radar/internal/domain"
)

// JobRepository define as operações de persistência e consulta de vagas.
type JobRepository interface {
	// Save insere uma nova vaga ou atualiza a data de última verificação se já existir.
	// Retorna isNew = true apenas se for um registro recém-criado.
	Save(ctx context.Context, job *domain.Job) (isNew bool, err error)

	// ListRecent retorna as oportunidades cadastradas mais recentes.
	ListRecent(ctx context.Context, limit int) ([]domain.Job, error)

	// GetByID busca uma oportunidade específica pelo ID.
	GetByID(ctx context.Context, id int64) (*domain.Job, error)

	// Count retorna a contagem total de vagas salvas no banco.
	Count(ctx context.Context) (int64, error)

	// Close encerra as conexões ativas do repositório.
	Close()
}
