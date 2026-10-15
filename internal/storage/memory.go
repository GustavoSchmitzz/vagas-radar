package storage

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/GustavoSchmitzz/vagas-radar/internal/domain"
)

// MemoryRepository implementa JobRepository em memória com thread-safety.
// Ideal para testes unitários rápidos e execução simulada.
type MemoryRepository struct {
	mu     sync.RWMutex
	jobs   map[string]domain.Job
	nextID int64
}

// NewMemoryRepository inicializa o repositório em memória.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		jobs:   make(map[string]domain.Job),
		nextID: 1,
	}
}

// Save persiste ou atualiza a vaga em memória.
func (m *MemoryRepository) Save(ctx context.Context, job *domain.Job) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if job.Hash == "" {
		job.ComputeHash()
	}

	existing, exists := m.jobs[job.Hash]
	now := time.Now()

	if exists {
		existing.UpdatedAt = now
		m.jobs[job.Hash] = existing
		job.ID = existing.ID
		return false, nil
	}

	job.ID = m.nextID
	m.nextID++
	job.CreatedAt = now
	job.UpdatedAt = now

	m.jobs[job.Hash] = *job
	return true, nil
}

// ListRecent retorna as oportunidades em memória ordenadas pelas mais recentes.
func (m *MemoryRepository) ListRecent(ctx context.Context, limit int) ([]domain.Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]domain.Job, 0, len(m.jobs))
	for _, j := range m.jobs {
		list = append(list, j)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].PublishedAt.After(list[j].PublishedAt)
	})

	if limit > 0 && len(list) > limit {
		list = list[:limit]
	}

	return list, nil
}

// GetByID busca uma vaga específica pelo ID em memória.
func (m *MemoryRepository) GetByID(ctx context.Context, id int64) (*domain.Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, j := range m.jobs {
		if j.ID == id {
			return &j, nil
		}
	}
	return nil, fmt.Errorf("vaga #%d não encontrada", id)
}

// Count retorna o total de registros armazenados.
func (m *MemoryRepository) Count(ctx context.Context) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return int64(len(m.jobs)), nil
}

// Close libera recursos (no-op para memória).
func (m *MemoryRepository) Close() {}
