package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// DatabaseConfig contém os parâmetros de conexão com o PostgreSQL.
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
	MaxConns int32
}

// DSN retorna a Connection String compatível com pgx / PostgreSQL.
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User,
		d.Password,
		d.Host,
		d.Port,
		d.Database,
		d.SSLMode,
	)
}

// FilterConfig define as regras ativas de triagem de vagas e processos seletivos.
type FilterConfig struct {
	// StrictInternshipOnly quando true rejeita qualquer vaga que não seja estágio/programa de estágio.
	// Quando false, permite que vagas de outras senioridades (Júnior, etc.) e CLT passem pelo filtro.
	StrictInternshipOnly bool

	// AllowCLT permite regime CLT caso o filtro estrito de estágio esteja desativado.
	AllowCLT bool

	// IncludeSelectionPrograms aceita programas corporativos e processos seletivos para estudantes.
	IncludeSelectionPrograms bool

	// Cidades presenciais ou híbridas aceitas (Cuiabá e Várzea Grande por padrão)
	TargetCities []string

	// Estado sede das vagas presenciais/híbridas
	TargetState string
}

// SchedulerConfig define a periodicidade das varreduras em segundo plano.
type SchedulerConfig struct {
	Interval       time.Duration
	RequestTimeout time.Duration
}

// Config agrega todas as configurações da aplicação.
type Config struct {
	Database    DatabaseConfig
	Filter      FilterConfig
	Scheduler   SchedulerConfig
	GitHubToken string
}

// LoadConfig carrega as configurações com valores padrão seguros, sobrescrevendo
// com variáveis de ambiente quando presentes.
func LoadConfig() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("POSTGRES_HOST", getEnv("PGHOST", "localhost")),
			Port:     getEnvAsInt("POSTGRES_PORT", getEnvAsInt("PGPORT", 5432)),
			User:     getEnv("POSTGRES_USER", getEnv("PGUSER", "postgres")),
			Password: getEnv("POSTGRES_PASSWORD", getEnv("PGPASSWORD", "")),
			Database: getEnv("POSTGRES_DB", getEnv("PGDATABASE", "vagas_radar")),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
			MaxConns: int32(getEnvAsInt("POSTGRES_MAX_CONNS", 5)),
		},
		Filter: FilterConfig{
			StrictInternshipOnly:     getEnvAsBool("STRICT_INTERNSHIP_ONLY", true),
			AllowCLT:                 getEnvAsBool("ALLOW_CLT", false),
			IncludeSelectionPrograms: getEnvAsBool("INCLUDE_SELECTION_PROGRAMS", true),
			TargetCities: []string{
				"cuiaba",
				"cuiabá",
				"varzea grande",
				"várzea grande",
			},
			TargetState: "MT",
		},
		Scheduler: SchedulerConfig{
			Interval:       getEnvAsDuration("SCAN_INTERVAL", 2*time.Hour),
			RequestTimeout: getEnvAsDuration("REQUEST_TIMEOUT", 30*time.Second),
		},
		GitHubToken: getEnv("GITHUB_TOKEN", ""),
	}
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && strings.TrimSpace(val) != "" {
		return strings.TrimSpace(val)
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	if valStr, ok := os.LookupEnv(key); ok {
		if val, err := strconv.Atoi(strings.TrimSpace(valStr)); err == nil {
			return val
		}
	}
	return defaultVal
}

func getEnvAsBool(key string, defaultVal bool) bool {
	if valStr, ok := os.LookupEnv(key); ok {
		valStr = strings.ToLower(strings.TrimSpace(valStr))
		if valStr == "true" || valStr == "1" || valStr == "yes" {
			return true
		}
		if valStr == "false" || valStr == "0" || valStr == "no" {
			return false
		}
	}
	return defaultVal
}

func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	if valStr, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(strings.TrimSpace(valStr)); err == nil {
			return d
		}
	}
	return defaultVal
}
