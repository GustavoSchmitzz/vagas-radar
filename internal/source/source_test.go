package source

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGitHubSource_Fetch(t *testing.T) {
	// Cria servidor HTTP simulando a resposta da API do GitHub
	mockResponse := `[
		{
			"id": 1001,
			"number": 42,
			"title": "[Remoto] Estágio em Desenvolvimento Go na FinTech Brasil",
			"body": "Buscamos estudante para atuar no desenvolvimento de microsserviços.",
			"html_url": "https://github.com/backend-br/vagas/issues/42",
			"created_at": "2026-10-15T12:00:00Z",
			"labels": [
				{"name": "Remoto"},
				{"name": "Estágio"}
			]
		},
		{
			"id": 1002,
			"number": 43,
			"title": "PR para adicionar funcionalidade",
			"body": "Ignorar",
			"html_url": "https://github.com/backend-br/vagas/pull/43",
			"created_at": "2026-10-15T12:30:00Z",
			"labels": []
		}
	]`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/backend-br/vagas/issues" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	src := NewGitHubSource("backend-br", "vagas", server.Client())
	src.SetBaseURL(server.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	jobs, err := src.Fetch(ctx)
	if err != nil {
		t.Fatalf("erro ao executar Fetch: %v", err)
	}

	// Deve ter ignorado o Pull Request e retornado apenas a issue
	if len(jobs) != 1 {
		t.Fatalf("esperava 1 vaga (ignorando PR), obteve %d", len(jobs))
	}

	job := jobs[0]
	if job.ExternalID != "1001" {
		t.Errorf("esperava ExternalID '1001', obteve '%s'", job.ExternalID)
	}
	if job.Company != "FinTech Brasil" {
		t.Errorf("esperava empresa 'FinTech Brasil', obteve '%s'", job.Company)
	}
	if job.URL != "https://github.com/backend-br/vagas/issues/42" {
		t.Errorf("esperava URL da issue, obteve '%s'", job.URL)
	}
	if job.Source != "github:backend-br/vagas" {
		t.Errorf("esperava Source 'github:backend-br/vagas', obteve '%s'", job.Source)
	}
}

func TestSelectionProgramSource_Fetch(t *testing.T) {
	mockFeed := `[
		{
			"id": "prog-2026",
			"title": "Programa de Estágio Tech 2026",
			"company": "Tech Academy",
			"description": "Inscrições abertas para estudantes de tecnologia em todo o país.",
			"url": "https://techacademy.com.br/estagio2026",
			"location": "Brasil",
			"work_mode": "Remoto"
		}
	]`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockFeed))
	}))
	defer server.Close()

	src := NewSelectionProgramSource("programas:tech", server.URL, server.Client())

	ctx := context.Background()
	jobs, err := src.Fetch(ctx)
	if err != nil {
		t.Fatalf("erro ao buscar feed: %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("esperava 1 programa coletado, obteve %d", len(jobs))
	}

	if jobs[0].Title != "Programa de Estágio Tech 2026" {
		t.Errorf("título inesperado: %s", jobs[0].Title)
	}
}

func TestRegistry(t *testing.T) {
	reg := NewRegistry()
	src1 := NewGitHubSource("frontendbr", "vagas", nil)
	src2 := NewSelectionProgramSource("programas", "", nil)

	reg.Register(src1)
	reg.Register(src2)

	all := reg.All()
	if len(all) != 2 {
		t.Errorf("esperava 2 fontes no registry, obteve %d", len(all))
	}
}

func TestRemotiveSource_Fetch(t *testing.T) {
	mockJSON := `{
		"jobs": [
			{
				"id": 5001,
				"url": "https://remotive.com/job/5001",
				"title": "Software Engineer Intern",
				"company_name": "Global Corp",
				"candidate_required_location": "Worldwide",
				"job_type": "internship",
				"publication_date": "2026-10-15T10:00:00",
				"description": "Internship opportunity for university students."
			},
			{
				"id": 5002,
				"url": "https://remotive.com/job/5002",
				"title": "Developer",
				"company_name": "Local US Corp",
				"candidate_required_location": "USA Only",
				"job_type": "full_time",
				"publication_date": "2026-10-15T10:00:00",
				"description": "US only."
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockJSON))
	}))
	defer server.Close()

	src := NewRemotiveSource(server.Client())
	src.SetAPIURL(server.URL)

	jobs, err := src.Fetch(context.Background())
	if err != nil {
		t.Fatalf("erro ao executar Fetch no Remotive: %v", err)
	}

	// Apenas a vaga Worldwide deve ser aceita (a vaga USA Only deve ser descartada pelo coletor)
	if len(jobs) != 1 {
		t.Fatalf("esperava 1 vaga filtrada pelo Remotive, obteve %d", len(jobs))
	}

	if jobs[0].Company != "Global Corp" {
		t.Errorf("empresa inesperada: %s", jobs[0].Company)
	}
}

func TestRSSSource_Fetch(t *testing.T) {
	mockXML := `<?xml version="1.0" encoding="UTF-8"?>
	<rss version="2.0">
	  <channel>
	    <title>Tech Remote Jobs</title>
	    <item>
	      <title>Tech Startup: Backend Software Intern</title>
	      <link>https://weworkremotely.com/jobs/123</link>
	      <description>Estágio para estudantes de graduação.</description>
	      <pubDate>Thu, 15 Oct 2026 12:00:00 GMT</pubDate>
	    </item>
	  </channel>
	</rss>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(mockXML))
	}))
	defer server.Close()

	src := NewRSSSource("weworkremotely", server.URL, server.Client())

	jobs, err := src.Fetch(context.Background())
	if err != nil {
		t.Fatalf("erro no RSS Fetch: %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("esperava 1 vaga coletada do RSS, obteve %d", len(jobs))
	}

	if jobs[0].Company != "Tech Startup" {
		t.Errorf("empresa incorreta extraída do RSS: %s", jobs[0].Company)
	}
	if jobs[0].Title != "Backend Software Intern" {
		t.Errorf("título incorreto extraído do RSS: %s", jobs[0].Title)
	}
}
