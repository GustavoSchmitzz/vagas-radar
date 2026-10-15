package runner

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/GustavoSchmitzz/vagas-radar/internal/config"
	"github.com/GustavoSchmitzz/vagas-radar/internal/domain"
	"github.com/GustavoSchmitzz/vagas-radar/internal/filter"
	"github.com/GustavoSchmitzz/vagas-radar/internal/notifier"
	"github.com/GustavoSchmitzz/vagas-radar/internal/source"
	"github.com/GustavoSchmitzz/vagas-radar/internal/storage"
)

type mockSource struct {
	name string
	jobs []domain.RawJob
	err  error
}

func (m *mockSource) Name() string { return m.name }
func (m *mockSource) Fetch(ctx context.Context) ([]domain.RawJob, error) {
	return m.jobs, m.err
}

type mockNotifier struct {
	mu           sync.Mutex
	notifiedJobs []*domain.Job
}

func (m *mockNotifier) NotifyNewJob(job *domain.Job) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notifiedJobs = append(m.notifiedJobs, job)
}

func (m *mockNotifier) PrintSummary(stats notifier.SummaryStats) {}

func TestRunner_RunOnce(t *testing.T) {
	ctx := context.Background()

	cfg := config.FilterConfig{
		StrictInternshipOnly:     true,
		IncludeSelectionPrograms: true,
		TargetCities:             []string{"cuiaba", "cuiabá", "varzea grande", "várzea grande"},
		TargetState:              "MT",
	}

	filterEngine := filter.NewEngine(cfg)
	repo := storage.NewMemoryRepository()
	notif := &mockNotifier{}

	reg := source.NewRegistry()

	// Fonte 1: com uma vaga válida e uma vaga fora do perfil
	src1 := &mockSource{
		name: "test:source1",
		jobs: []domain.RawJob{
			{
				Title:       "Estágio em Engenharia de Software",
				Company:     "Empresa Local Cuiabá",
				URL:         "https://example.com/job/1",
				LocationRaw: "Cuiabá, MT",
				WorkModeRaw: "Presencial",
				PublishedAt: time.Now(),
			},
			{
				Title:       "Desenvolvedor Sênior Java",
				Company:     "Big Bank",
				URL:         "https://example.com/job/2",
				LocationRaw: "São Paulo, SP",
				WorkModeRaw: "Presencial",
				PublishedAt: time.Now(),
			},
		},
	}

	// Fonte 2: fonte com falha simulada (não deve travar o runner)
	src2 := &mockSource{
		name: "test:failing",
		err:  errors.New("conexão recusada"),
	}

	reg.Register(src1)
	reg.Register(src2)

	runner := NewRunner(reg, filterEngine, repo, notif, 2)

	stats, err := runner.RunOnce(ctx)
	if err != nil {
		t.Fatalf("RunOnce falhou inesperadamente: %v", err)
	}

	if stats.TotalSources != 2 {
		t.Errorf("esperava 2 fontes consultadas, obteve %d", stats.TotalSources)
	}
	if stats.FailedSources != 1 {
		t.Errorf("esperava 1 falha isolada na fonte 2, obteve %d", stats.FailedSources)
	}
	if stats.TotalEvaluated != 2 {
		t.Errorf("esperava 2 vagas avaliadas, obteve %d", stats.TotalEvaluated)
	}
	if stats.TotalAccepted != 1 {
		t.Errorf("esperava 1 vaga aceita (apenas estágio em Cuiabá), obteve %d", stats.TotalAccepted)
	}
	if stats.TotalNewSaved != 1 {
		t.Errorf("esperava 1 nova vaga salva, obteve %d", stats.TotalNewSaved)
	}

	if len(notif.notifiedJobs) != 1 {
		t.Fatalf("esperava 1 notificação emitida, obteve %d", len(notif.notifiedJobs))
	}
	if notif.notifiedJobs[0].Title != "Estágio em Engenharia de Software" {
		t.Errorf("título notificado incorreto: %s", notif.notifiedJobs[0].Title)
	}
}
