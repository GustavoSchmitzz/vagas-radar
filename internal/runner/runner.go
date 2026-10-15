package runner

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/GustavoSchmitzz/vagas-radar/internal/domain"
	"github.com/GustavoSchmitzz/vagas-radar/internal/filter"
	"github.com/GustavoSchmitzz/vagas-radar/internal/notifier"
	"github.com/GustavoSchmitzz/vagas-radar/internal/source"
	"github.com/GustavoSchmitzz/vagas-radar/internal/storage"
)

// Runner gerencia a execução orquestrada e concorrente de varredura.
type Runner struct {
	registry     *source.Registry
	filterEngine *filter.Engine
	repository   storage.JobRepository
	notifier     notifier.Notifier
	maxWorkers   int
}

// NewRunner instancia um novo orquestrador.
func NewRunner(
	registry *source.Registry,
	filterEngine *filter.Engine,
	repository storage.JobRepository,
	notifier notifier.Notifier,
	maxWorkers int,
) *Runner {
	if maxWorkers <= 0 {
		maxWorkers = 4
	}
	return &Runner{
		registry:     registry,
		filterEngine: filterEngine,
		repository:   repository,
		notifier:     notifier,
		maxWorkers:   maxWorkers,
	}
}

// RunOnce executa um ciclo completo de varredura:
// 1. Consulta fontes registradas em paralelo com limite de concorrência.
// 2. Submete cada vaga ao pipeline de validação (estágio, senioridade, localização).
// 3. Persiste no banco com deduplicação.
// 4. Notifica oportunidades inéditas.
func (r *Runner) RunOnce(ctx context.Context) (notifier.SummaryStats, error) {
	sources := r.registry.All()
	stats := notifier.SummaryStats{
		TotalSources: len(sources),
	}

	if len(sources) == 0 {
		return stats, nil
	}

	// Semáforo para controlar concorrência máxima de requisições de rede
	sem := make(chan struct{}, r.maxWorkers)
	var wg sync.WaitGroup
	var mu sync.Mutex

	collectedJobs := make([]domain.RawJob, 0)

	for _, src := range sources {
		wg.Add(1)
		go func(s source.JobSource) {
			defer wg.Done()

			sem <- struct{}{}        // adquire token
			defer func() { <-sem }() // libera token

			srcCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
			defer cancel()

			raws, err := s.Fetch(srcCtx)
			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				stats.FailedSources++
				log.Printf("[ALERTA] Falha na fonte %s: %v", s.Name(), err)
				return
			}

			collectedJobs = append(collectedJobs, raws...)
		}(src)
	}

	wg.Wait()

	// Avalia e persiste cada oportunidade coletada
	for _, raw := range collectedJobs {
		stats.TotalEvaluated++

		eval := r.filterEngine.Evaluate(raw)
		if !eval.Accepted || eval.Job == nil {
			continue
		}

		stats.TotalAccepted++

		isNew, err := r.repository.Save(ctx, eval.Job)
		if err != nil {
			log.Printf("[ERRO] Falha ao persistir vaga '%s': %v", eval.Job.Title, err)
			continue
		}

		if isNew {
			stats.TotalNewSaved++
			if r.notifier != nil {
				r.notifier.NotifyNewJob(eval.Job)
			}
		}
	}

	if r.notifier != nil {
		r.notifier.PrintSummary(stats)
	}

	return stats, nil
}

// StartDaemon inicia a execução em loop contínuo pelo intervalo estipulado.
func (r *Runner) StartDaemon(ctx context.Context, interval time.Duration) error {
	log.Printf("[DAEMON] Radar de Vagas ativo. Intervalo entre buscas: %v", interval)

	// Primeira execução imediata ao iniciar o serviço
	log.Println("[DAEMON] Iniciando varredura inicial...")
	if _, err := r.RunOnce(ctx); err != nil {
		log.Printf("[DAEMON] Erro na varredura inicial: %v", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[DAEMON] Sinal de encerramento recebido. Finalizando serviço com segurança...")
			return nil
		case t := <-ticker.C:
			log.Printf("[DAEMON] Iniciando nova varredura agendada às %s...", t.Format("15:04:05"))
			if _, err := r.RunOnce(ctx); err != nil {
				log.Printf("[DAEMON] Erro durante a varredura: %v", err)
			}
		}
	}
}
