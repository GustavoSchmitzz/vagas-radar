package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/GustavoSchmitzz/vagas-radar/internal/config"
	"github.com/GustavoSchmitzz/vagas-radar/internal/domain"
)

// PostgresRepository implementa JobRepository utilizando o driver oficial pgx com connection pool.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository inicializa o pool de conexões com o PostgreSQL e executa migrações automáticas.
func NewPostgresRepository(ctx context.Context, cfg config.DatabaseConfig) (*PostgresRepository, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("falha ao interpretar DSN do banco: %w", err)
	}

	if cfg.MaxConns > 0 {
		poolConfig.MaxConns = cfg.MaxConns
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar no PostgreSQL: %w", err)
	}

	// Testa a conexão
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("PostgreSQL inacessível (%s:%d): %w", cfg.Host, cfg.Port, err)
	}

	repo := &PostgresRepository{pool: pool}
	if err := repo.migrate(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("falha ao aplicar migrações iniciais: %w", err)
	}

	return repo, nil
}

// migrate cria as tabelas e índices necessários caso não existam.
func (r *PostgresRepository) migrate(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS jobs (
		id BIGSERIAL PRIMARY KEY,
		hash VARCHAR(64) UNIQUE NOT NULL,
		title TEXT NOT NULL,
		company TEXT NOT NULL,
		description TEXT,
		url TEXT NOT NULL,
		source VARCHAR(64) NOT NULL,
		city VARCHAR(128),
		state VARCHAR(32),
		country VARCHAR(64),
		location_raw TEXT,
		work_mode VARCHAR(32) NOT NULL,
		opportunity_type VARCHAR(32) NOT NULL,
		seniority VARCHAR(32) NOT NULL,
		contract_type VARCHAR(32) NOT NULL,
		published_at TIMESTAMPTZ NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_jobs_opportunity_type ON jobs(opportunity_type);
	CREATE INDEX IF NOT EXISTS idx_jobs_work_mode ON jobs(work_mode);
	CREATE INDEX IF NOT EXISTS idx_jobs_published_at ON jobs(published_at DESC);
	`
	_, err := r.pool.Exec(ctx, query)
	return err
}

// Save persiste a oportunidade de forma atômica e com deduplicação por hash.
// Retorna isNew = true se a oportunidade foi cadastrada pela primeira vez.
func (r *PostgresRepository) Save(ctx context.Context, job *domain.Job) (bool, error) {
	if job.Hash == "" {
		job.ComputeHash()
	}

	query := `
	INSERT INTO jobs (
		hash, title, company, description, url, source,
		city, state, country, location_raw,
		work_mode, opportunity_type, seniority, contract_type,
		published_at, created_at, updated_at
	) VALUES (
		$1, $2, $3, $4, $5, $6,
		$7, $8, $9, $10,
		$11, $12, $13, $14,
		$15, NOW(), NOW()
	)
	ON CONFLICT (hash) DO UPDATE SET
		updated_at = NOW()
	RETURNING id, (xmax = 0) AS is_new;
	`

	var id int64
	var isNew bool

	err := r.pool.QueryRow(ctx, query,
		job.Hash,
		job.Title,
		job.Company,
		job.Description,
		job.URL,
		job.Source,
		job.Location.City,
		job.Location.State,
		job.Location.Country,
		job.Location.Raw,
		string(job.WorkMode),
		string(job.OpportunityType),
		string(job.Seniority),
		string(job.ContractType),
		job.PublishedAt,
	).Scan(&id, &isNew)

	if err != nil {
		return false, fmt.Errorf("erro ao salvar vaga no PostgreSQL: %w", err)
	}

	job.ID = id
	return isNew, nil
}

// ListRecent retorna as últimas N oportunidades salvas, ordenadas pela data de publicação.
func (r *PostgresRepository) ListRecent(ctx context.Context, limit int) ([]domain.Job, error) {
	if limit <= 0 {
		limit = 20
	}

	query := `
	SELECT
		id, hash, title, company, description, url, source,
		city, state, country, location_raw,
		work_mode, opportunity_type, seniority, contract_type,
		published_at, created_at, updated_at
	FROM jobs
	ORDER BY published_at DESC, id DESC
	LIMIT $1;
	`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("falha ao consultar vagas recentes: %w", err)
	}
	defer rows.Close()

	var jobs []domain.Job
	for rows.Next() {
		var j domain.Job
		var workMode, oppType, seniority, contractType string

		err := rows.Scan(
			&j.ID,
			&j.Hash,
			&j.Title,
			&j.Company,
			&j.Description,
			&j.URL,
			&j.Source,
			&j.Location.City,
			&j.Location.State,
			&j.Location.Country,
			&j.Location.Raw,
			&workMode,
			&oppType,
			&seniority,
			&contractType,
			&j.PublishedAt,
			&j.CreatedAt,
			&j.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler registro de vaga: %w", err)
		}

		j.WorkMode = domain.WorkMode(workMode)
		j.OpportunityType = domain.OpportunityType(oppType)
		j.Seniority = domain.Seniority(seniority)
		j.ContractType = domain.ContractType(contractType)

		jobs = append(jobs, j)
	}

	return jobs, nil
}

// GetByID busca uma oportunidade específica pelo ID no PostgreSQL.
func (r *PostgresRepository) GetByID(ctx context.Context, id int64) (*domain.Job, error) {
	query := `
	SELECT
		id, hash, title, company, description, url, source,
		city, state, country, location_raw,
		work_mode, opportunity_type, seniority, contract_type,
		published_at, created_at, updated_at
	FROM jobs
	WHERE id = $1;
	`

	var j domain.Job
	var workMode, oppType, seniority, contractType string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&j.ID,
		&j.Hash,
		&j.Title,
		&j.Company,
		&j.Description,
		&j.URL,
		&j.Source,
		&j.Location.City,
		&j.Location.State,
		&j.Location.Country,
		&j.Location.Raw,
		&workMode,
		&oppType,
		&seniority,
		&contractType,
		&j.PublishedAt,
		&j.CreatedAt,
		&j.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("vaga #%d não encontrada: %w", id, err)
	}

	j.WorkMode = domain.WorkMode(workMode)
	j.OpportunityType = domain.OpportunityType(oppType)
	j.Seniority = domain.Seniority(seniority)
	j.ContractType = domain.ContractType(contractType)

	return &j, nil
}

// Count retorna o número total de oportunidades cadastradas.
func (r *PostgresRepository) Count(ctx context.Context) (int64, error) {
	var total int64
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM jobs").Scan(&total)
	return total, err
}

// Close encerra as conexões do pool com o PostgreSQL.
func (r *PostgresRepository) Close() {
	if r.pool != nil {
		r.pool.Close()
	}
}
