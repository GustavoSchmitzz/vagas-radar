package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfigDefaults(t *testing.T) {
	// Garante que não há variáveis de ambiente interferindo
	os.Unsetenv("POSTGRES_HOST")
	os.Unsetenv("STRICT_INTERNSHIP_ONLY")
	os.Unsetenv("SCAN_INTERVAL")

	cfg := LoadConfig()

	if cfg.Database.Host != "localhost" {
		t.Errorf("esperava host default 'localhost', obteve %s", cfg.Database.Host)
	}

	if cfg.Database.Port != 5432 {
		t.Errorf("esperava port default 5432, obteve %d", cfg.Database.Port)
	}

	if !cfg.Filter.StrictInternshipOnly {
		t.Errorf("esperava StrictInternshipOnly como true por padrão")
	}

	if cfg.Scheduler.Interval != 2*time.Hour {
		t.Errorf("esperava Interval default de 2h, obteve %v", cfg.Scheduler.Interval)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	os.Setenv("POSTGRES_HOST", "db.internal")
	os.Setenv("POSTGRES_PORT", "5433")
	os.Setenv("STRICT_INTERNSHIP_ONLY", "false")
	os.Setenv("ALLOW_CLT", "true")
	os.Setenv("SCAN_INTERVAL", "30m")
	defer func() {
		os.Unsetenv("POSTGRES_HOST")
		os.Unsetenv("POSTGRES_PORT")
		os.Unsetenv("STRICT_INTERNSHIP_ONLY")
		os.Unsetenv("ALLOW_CLT")
		os.Unsetenv("SCAN_INTERVAL")
	}()

	cfg := LoadConfig()

	if cfg.Database.Host != "db.internal" {
		t.Errorf("esperava host 'db.internal', obteve %s", cfg.Database.Host)
	}

	if cfg.Database.Port != 5433 {
		t.Errorf("esperava port 5433, obteve %d", cfg.Database.Port)
	}

	if cfg.Filter.StrictInternshipOnly != false {
		t.Errorf("esperava StrictInternshipOnly falso via env")
	}

	if cfg.Filter.AllowCLT != true {
		t.Errorf("esperava AllowCLT verdadeiro via env")
	}

	if cfg.Scheduler.Interval != 30*time.Minute {
		t.Errorf("esperava Interval de 30m, obteve %v", cfg.Scheduler.Interval)
	}
}
