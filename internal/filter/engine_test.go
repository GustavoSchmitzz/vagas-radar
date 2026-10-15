package filter

import (
	"testing"
	"time"

	"github.com/GustavoSchmitzz/vagas-radar/internal/config"
	"github.com/GustavoSchmitzz/vagas-radar/internal/domain"
)

func defaultTestConfig() config.FilterConfig {
	return config.FilterConfig{
		StrictInternshipOnly:     true,
		AllowCLT:                 false,
		IncludeSelectionPrograms: true,
		TargetCities: []string{
			"cuiaba",
			"cuiabá",
			"varzea grande",
			"várzea grande",
		},
		TargetState: "MT",
	}
}

func TestEngine_Evaluate(t *testing.T) {
	tests := []struct {
		name                 string
		cfg                  config.FilterConfig
		raw                  domain.RawJob
		wantAccepted         bool
		wantOpportunityType  domain.OpportunityType
		wantWorkMode         domain.WorkMode
		wantSeniority        domain.Seniority
	}{
		{
			name: "Estágio Backend Remoto aceito",
			cfg:  defaultTestConfig(),
			raw: domain.RawJob{
				Title:       "Estágio em Desenvolvimento Backend (Go/Python)",
				Company:     "Fintech Global",
				Description: "Vaga 100% remota para estudantes de Ciência da Computação e correlatas.",
				URL:         "https://example.com/vagas/1",
				LocationRaw: "Remoto - Brasil",
				WorkModeRaw: "Remoto",
				Source:      "test",
				PublishedAt: time.Now(),
			},
			wantAccepted:        true,
			wantOpportunityType: domain.OpportunityInternship,
			wantWorkMode:        domain.WorkModeRemote,
			wantSeniority:       domain.SeniorityInternship,
		},
		{
			name: "Processo Seletivo / Programa de Estágio aceito",
			cfg:  defaultTestConfig(),
			raw: domain.RawJob{
				Title:       "Programa de Estágio Tech 2026 - Inscrições Abertas",
				Company:     "Big Tech Brasil",
				Description: "Venha participar do nosso processo seletivo para estágio em engenharia de software.",
				URL:         "https://example.com/vagas/2",
				LocationRaw: "Brasil",
				WorkModeRaw: "Remoto",
				Source:      "test",
				PublishedAt: time.Now(),
			},
			wantAccepted:        true,
			wantOpportunityType: domain.OpportunitySelectionProcess,
			wantWorkMode:        domain.WorkModeRemote,
			wantSeniority:       domain.SeniorityInternship,
		},
		{
			name: "Estágio Presencial em Cuiabá aceito",
			cfg:  defaultTestConfig(),
			raw: domain.RawJob{
				Title:       "Estagiário de Engenharia de Software",
				Company:     "AgroTech Cuiabá",
				Description: "Atividades presenciais no escritório central em Cuiabá - MT.",
				URL:         "https://example.com/vagas/3",
				LocationRaw: "Cuiabá, MT",
				WorkModeRaw: "Presencial",
				Source:      "test",
				PublishedAt: time.Now(),
			},
			wantAccepted:        true,
			wantOpportunityType: domain.OpportunityInternship,
			wantWorkMode:        domain.WorkModeOnSite,
			wantSeniority:       domain.SeniorityInternship,
		},
		{
			name: "Estágio Híbrido em Várzea Grande aceito",
			cfg:  defaultTestConfig(),
			raw: domain.RawJob{
				Title:       "Estágio em QA e Testes de Software",
				Company:     "Logística MT",
				Description: "Modelo híbrido com 2 dias em Várzea Grande - MT.",
				URL:         "https://example.com/vagas/4",
				LocationRaw: "Várzea Grande - MT",
				WorkModeRaw: "Híbrido",
				Source:      "test",
				PublishedAt: time.Now(),
			},
			wantAccepted:        true,
			wantOpportunityType: domain.OpportunityInternship,
			wantWorkMode:        domain.WorkModeHybrid,
			wantSeniority:       domain.SeniorityInternship,
		},
		{
			name: "Estágio Presencial em São Paulo REJEITADO",
			cfg:  defaultTestConfig(),
			raw: domain.RawJob{
				Title:       "Estágio em Desenvolvimento Frontend",
				Company:     "Paulista Tech",
				Description: "Vaga 100% presencial na Av. Paulista, São Paulo - SP.",
				URL:         "https://example.com/vagas/5",
				LocationRaw: "São Paulo, SP",
				WorkModeRaw: "Presencial",
				Source:      "test",
				PublishedAt: time.Now(),
			},
			wantAccepted: false,
		},
		{
			name: "Estágio Híbrido em Curitiba REJEITADO",
			cfg:  defaultTestConfig(),
			raw: domain.RawJob{
				Title:       "Estágio Mobile (Flutter/React Native)",
				Company:     "Sul Apps",
				Description: "Modelo híbrido com presença semanal em Curitiba - PR.",
				URL:         "https://example.com/vagas/6",
				LocationRaw: "Curitiba, PR",
				WorkModeRaw: "Híbrido",
				Source:      "test",
				PublishedAt: time.Now(),
			},
			wantAccepted: false,
		},
		{
			name: "Vaga Remota com restrição regional impeditiva REJEITADA",
			cfg:  defaultTestConfig(),
			raw: domain.RawJob{
				Title:       "Estágio em DevOps - Trabalho Remoto",
				Company:     "Gov SP",
				Description: "Vaga remota apenas para residentes no estado de São Paulo.",
				URL:         "https://example.com/vagas/7",
				LocationRaw: "Remoto",
				WorkModeRaw: "Remoto",
				Source:      "test",
				PublishedAt: time.Now(),
			},
			wantAccepted: false,
		},
		{
			name: "Vaga Desenvolvedor Júnior com filtro estrito REJEITADA",
			cfg:  defaultTestConfig(),
			raw: domain.RawJob{
				Title:       "Desenvolvedor Backend Júnior (Go/PostgreSQL)",
				Company:     "Tech Corp",
				Description: "Contratação CLT para início imediato.",
				URL:         "https://example.com/vagas/8",
				LocationRaw: "Remoto",
				WorkModeRaw: "Remoto",
				Source:      "test",
				PublishedAt: time.Now(),
			},
			wantAccepted: false,
		},
		{
			name: "Vaga Desenvolvedor Júnior com filtro relaxado ACEITA",
			cfg: func() config.FilterConfig {
				c := defaultTestConfig()
				c.StrictInternshipOnly = false
				c.AllowCLT = true
				return c
			}(),
			raw: domain.RawJob{
				Title:       "Desenvolvedor Backend Júnior",
				Company:     "Tech Corp",
				Description: "Vaga remota CLT nível júnior.",
				URL:         "https://example.com/vagas/9",
				LocationRaw: "Remoto",
				WorkModeRaw: "Remoto",
				Source:      "test",
				PublishedAt: time.Now(),
			},
			wantAccepted:        true,
			wantOpportunityType: domain.OpportunityGeneralJob,
			wantWorkMode:        domain.WorkModeRemote,
			wantSeniority:       domain.SeniorityJunior,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine(tt.cfg)
			res := engine.Evaluate(tt.raw)

			if res.Accepted != tt.wantAccepted {
				t.Fatalf("esperava Accepted=%v, obteve %v (motivo: %s)", tt.wantAccepted, res.Accepted, res.Reason)
			}

			if tt.wantAccepted {
				if res.Job == nil {
					t.Fatalf("esperava Job instanciado, veio nil")
				}
				if res.Job.OpportunityType != tt.wantOpportunityType {
					t.Errorf("esperava OpportunityType=%v, obteve %v", tt.wantOpportunityType, res.Job.OpportunityType)
				}
				if res.Job.WorkMode != tt.wantWorkMode {
					t.Errorf("esperava WorkMode=%v, obteve %v", tt.wantWorkMode, res.Job.WorkMode)
				}
				if res.Job.Seniority != tt.wantSeniority {
					t.Errorf("esperava Seniority=%v, obteve %v", tt.wantSeniority, res.Job.Seniority)
				}
				if res.Job.Hash == "" {
					t.Errorf("esperava Hash preenchido")
				}
			}
		})
	}
}
