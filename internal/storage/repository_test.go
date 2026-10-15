package storage

import (
	"context"
	"testing"
	"time"

	"github.com/GustavoSchmitzz/vagas-radar/internal/domain"
)

func TestMemoryRepository_SaveAndDeduplicate(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()

	job := &domain.Job{
		Title:           "Estágio em Engenharia de Software",
		Company:         "AgroTech Cuiabá",
		URL:             "https://example.com/vagas/100",
		Source:          "test",
		OpportunityType: domain.OpportunityInternship,
		WorkMode:        domain.WorkModeOnSite,
		Seniority:       domain.SeniorityInternship,
		ContractType:    domain.ContractInternship,
		PublishedAt:     time.Now().Add(-1 * time.Hour),
	}

	// 1ª inserção: deve ser considerada nova vaga
	isNew, err := repo.Save(ctx, job)
	if err != nil {
		t.Fatalf("erro ao salvar primeira vez: %v", err)
	}
	if !isNew {
		t.Errorf("esperava isNew=true na primeira inserção")
	}
	if job.ID <= 0 {
		t.Errorf("esperava ID válido atribuído à vaga")
	}

	total, err := repo.Count(ctx)
	if err != nil || total != 1 {
		t.Errorf("esperava 1 vaga cadastrada, obteve %d (err: %v)", total, err)
	}

	// 2ª inserção da mesma vaga (mesmo hash): deve ser detectada como existente
	initialID := job.ID
	isNewDuplicate, err := repo.Save(ctx, job)
	if err != nil {
		t.Fatalf("erro ao salvar duplicada: %v", err)
	}
	if isNewDuplicate {
		t.Errorf("esperava isNew=false para vaga duplicada")
	}
	if job.ID != initialID {
		t.Errorf("esperava ID mantido %d, obteve %d", initialID, job.ID)
	}

	totalAfter, err := repo.Count(ctx)
	if err != nil || totalAfter != 1 {
		t.Errorf("total de vagas não deveria ter aumentado na duplicata, obteve %d", totalAfter)
	}
}

func TestMemoryRepository_ListRecent(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()

	now := time.Now()
	jobs := []*domain.Job{
		{
			Title:       "Vaga Antiga",
			Company:     "Empresa A",
			URL:         "https://example.com/1",
			Source:      "test",
			PublishedAt: now.Add(-2 * time.Hour),
		},
		{
			Title:       "Vaga Nova",
			Company:     "Empresa B",
			URL:         "https://example.com/2",
			Source:      "test",
			PublishedAt: now,
		},
	}

	for _, j := range jobs {
		if _, err := repo.Save(ctx, j); err != nil {
			t.Fatalf("falha ao salvar: %v", err)
		}
	}

	recent, err := repo.ListRecent(ctx, 1)
	if err != nil {
		t.Fatalf("falha ao listar recentes: %v", err)
	}

	if len(recent) != 1 {
		t.Fatalf("esperava 1 vaga no limite, obteve %d", len(recent))
	}

	if recent[0].Title != "Vaga Nova" {
		t.Errorf("esperava 'Vaga Nova' como primeiro resultado, obteve '%s'", recent[0].Title)
	}
}
