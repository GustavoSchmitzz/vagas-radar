package source

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/GustavoSchmitzz/vagas-radar/internal/domain"
)

// GitHubIssue representa a estrutura reduzida de uma issue retornada pela API do GitHub.
type GitHubIssue struct {
	ID        int64     `json:"id"`
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	HTMLURL   string    `json:"html_url"`
	CreatedAt time.Time `json:"created_at"`
	Labels    []struct {
		Name string `json:"name"`
	} `json:"labels"`
}

// GitHubSource coleta vagas publicadas como issues em repositórios abertos da comunidade tech brasileira.
type GitHubSource struct {
	name       string
	owner      string
	repo       string
	apiBaseURL string
	client     *http.Client
}

// NewGitHubSource cria uma fonte para um repositório GitHub público (ex: "backend-br", "vagas").
func NewGitHubSource(owner, repo string, client *http.Client) *GitHubSource {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return &GitHubSource{
		name:       fmt.Sprintf("github:%s/%s", owner, repo),
		owner:      owner,
		repo:       repo,
		apiBaseURL: "https://api.github.com",
		client:     client,
	}
}

// SetBaseURL permite sobrescrever a URL base para testes unitários com httptest.Server.
func (s *GitHubSource) SetBaseURL(url string) {
	s.apiBaseURL = strings.TrimRight(url, "/")
}

// Name retorna o identificador da fonte.
func (s *GitHubSource) Name() string {
	return s.name
}

// Fetch consulta a API do GitHub e retorna as últimas vagas abertas no repositório.
func (s *GitHubSource) Fetch(ctx context.Context) ([]domain.RawJob, error) {
	endpoint := fmt.Sprintf("%s/repos/%s/%s/issues?state=open&per_page=30", s.apiBaseURL, s.owner, s.repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição para %s: %w", s.name, err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "vagas-radar/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("falha ao consultar %s: %w", s.name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("resposta inesperada da API do GitHub (%s): status %d", s.name, resp.StatusCode)
	}

	var issues []GitHubIssue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, fmt.Errorf("erro ao decodificar JSON de %s: %w", s.name, err)
	}

	rawJobs := make([]domain.RawJob, 0, len(issues))
	for _, issue := range issues {
		// Pula pull requests retornados na listagem de issues
		if strings.Contains(issue.HTMLURL, "/pull/") {
			continue
		}

		labelsStr := make([]string, 0, len(issue.Labels))
		for _, l := range issue.Labels {
			labelsStr = append(labelsStr, l.Name)
		}
		joinedLabels := strings.Join(labelsStr, " | ")

		company, cleanTitle := s.parseCompanyAndTitle(issue.Title)

		rawJobs = append(rawJobs, domain.RawJob{
			ExternalID:  fmt.Sprintf("%d", issue.ID),
			Title:       cleanTitle,
			Company:     company,
			Description: issue.Body,
			URL:         issue.HTMLURL,
			LocationRaw: joinedLabels,
			WorkModeRaw: joinedLabels,
			Source:      s.name,
			PublishedAt: issue.CreatedAt,
			Extra: map[string]string{
				"issue_number": fmt.Sprintf("%d", issue.Number),
				"labels":       joinedLabels,
			},
		})
	}

	return rawJobs, nil
}

// parseCompanyAndTitle tenta extrair a empresa e o cargo de formatos comuns:
// "[São Paulo/Remoto] [Estágio] Desenvolvedor Backend na Empresa X"
// "Empresa X contrata: Estagiário de Software"
func (s *GitHubSource) parseCompanyAndTitle(rawTitle string) (company, title string) {
	title = strings.TrimSpace(rawTitle)
	company = "Confidencial / Empresa Tech"

	if idx := strings.LastIndex(strings.ToLower(title), " na "); idx != -1 {
		company = strings.TrimSpace(title[idx+4:])
	} else if idx := strings.LastIndex(strings.ToLower(title), " no "); idx != -1 {
		company = strings.TrimSpace(title[idx+4:])
	} else if idx := strings.LastIndex(strings.ToLower(title), " @ "); idx != -1 {
		company = strings.TrimSpace(title[idx+3:])
	}

	return company, title
}
