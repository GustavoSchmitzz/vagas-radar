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

// RemotiveResponse representa a resposta da API pública do Remotive.
type RemotiveResponse struct {
	Jobs []RemotiveJob `json:"jobs"`
}

// RemotiveJob representa os campos de uma oportunidade no Remotive.
type RemotiveJob struct {
	ID                        int64  `json:"id"`
	URL                       string `json:"url"`
	Title                     string `json:"title"`
	CompanyName               string `json:"company_name"`
	CandidateRequiredLocation string `json:"candidate_required_location"`
	JobType                   string `json:"job_type"`
	PublicationDate           string `json:"publication_date"`
	Description               string `json:"description"`
}

// RemotiveSource consulta a API pública oficial do Remotive para vagas remotas.
type RemotiveSource struct {
	name   string
	apiURL string
	client *http.Client
}

// NewRemotiveSource cria uma nova fonte para vagas de tecnologia do Remotive.
func NewRemotiveSource(client *http.Client) *RemotiveSource {
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	return &RemotiveSource{
		name:   "remotive:software-dev",
		apiURL: "https://remotive.com/api/remote-jobs?category=software-dev&limit=40",
		client: client,
	}
}

// SetAPIURL permite customizar o endpoint (útil para testes unitários).
func (s *RemotiveSource) SetAPIURL(url string) {
	s.apiURL = url
}

// Name retorna o identificador legível da fonte.
func (s *RemotiveSource) Name() string {
	return s.name
}

// Fetch consulta a API do Remotive e normaliza as oportunidades.
func (s *RemotiveSource) Fetch(ctx context.Context) ([]domain.RawJob, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição para Remotive: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "vagas-radar/1.0 (Mozilla/5.0)")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("falha ao consultar API Remotive: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("resposta inesperada da API Remotive: status %d", resp.StatusCode)
	}

	var data RemotiveResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("falha ao decodificar JSON do Remotive: %w", err)
	}

	rawJobs := make([]domain.RawJob, 0, len(data.Jobs))
	for _, j := range data.Jobs {
		loc := strings.TrimSpace(j.CandidateRequiredLocation)

		// Filtra apenas vagas que aceitam candidatos do Brasil / LATAM / Worldwide
		locLower := strings.ToLower(loc)
		isEligible := locLower == "worldwide" ||
			locLower == "anywhere" ||
			strings.Contains(locLower, "brazil") ||
			strings.Contains(locLower, "brasil") ||
			strings.Contains(locLower, "latam") ||
			strings.Contains(locLower, "south america") ||
			strings.Contains(locLower, "americas")

		if !isEligible {
			continue
		}

		pubDate, _ := time.Parse("2006-01-02T15:04:05", j.PublicationDate)
		if pubDate.IsZero() {
			pubDate = time.Now()
		}

		rawJobs = append(rawJobs, domain.RawJob{
			ExternalID:  fmt.Sprintf("%d", j.ID),
			Title:       j.Title,
			Company:     j.CompanyName,
			Description: j.Description,
			URL:         j.URL,
			LocationRaw: "Remoto - " + loc,
			WorkModeRaw: "Remoto",
			Source:      s.name,
			PublishedAt: pubDate,
			Extra: map[string]string{
				"required_location": loc,
				"job_type":          j.JobType,
			},
		})
	}

	return rawJobs, nil
}
