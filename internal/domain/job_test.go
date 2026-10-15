package domain

import (
	"testing"
)

func TestGenerateHash(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		company  string
		title    string
		url      string
		sameAs   func() string
		diffFrom func() string
	}{
		{
			name:    "Hash determinístico por URL",
			source:  "github",
			company: "Empresa X",
			title:   "Estágio Backend",
			url:     "https://github.com/backend-br/vagas/issues/123",
			sameAs: func() string {
				// Mesma URL com espaços ou maiúsculas deve gerar o mesmo hash
				return GenerateHash("github", "Empresa Y", "Outro Titulo", " https://github.com/backend-br/vagas/issues/123 ")
			},
			diffFrom: func() string {
				return GenerateHash("github", "Empresa X", "Estágio Backend", "https://github.com/backend-br/vagas/issues/124")
			},
		},
		{
			name:    "Hash fallback quando URL está vazia",
			source:  "portal",
			company: "Tech Corp",
			title:   "Estagiário de QA",
			url:     "",
			sameAs: func() string {
				return GenerateHash("portal", " Tech Corp ", " estagiário de qa ", "")
			},
			diffFrom: func() string {
				return GenerateHash("portal", "Tech Corp", "Estagiário Frontend", "")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h1 := GenerateHash(tt.source, tt.company, tt.title, tt.url)
			if h1 == "" {
				t.Fatalf("esperava hash não-vazio")
			}

			if tt.sameAs != nil {
				hSame := tt.sameAs()
				if h1 != hSame {
					t.Errorf("esperava hashes idênticos, obteve %s vs %s", h1, hSame)
				}
			}

			if tt.diffFrom != nil {
				hDiff := tt.diffFrom()
				if h1 == hDiff {
					t.Errorf("esperava hashes diferentes, obteve %s igual", h1)
				}
			}
		})
	}
}

func TestJobComputeHash(t *testing.T) {
	job := &Job{
		Source:  "github",
		Company: "Startup MT",
		Title:   "Estágio em Engenharia de Software",
		URL:     "https://exemplo.com/vaga/1",
	}

	job.ComputeHash()
	if job.Hash == "" {
		t.Errorf("ComputeHash deveria ter gerado um hash não vazio")
	}

	expected := GenerateHash(job.Source, job.Company, job.Title, job.URL)
	if job.Hash != expected {
		t.Errorf("esperava hash %s, obteve %s", expected, job.Hash)
	}
}
