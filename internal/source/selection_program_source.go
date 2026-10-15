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

// ProgramItem representa o esquema de uma oportunidade vinda de feeds públicos de programas de estágio.
type ProgramItem struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Company     string    `json:"company"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	Location    string    `json:"location"`
	WorkMode    string    `json:"work_mode"`
	PublishedAt time.Time `json:"published_at"`
}

// SelectionProgramSource consulta repositórios de processos seletivos e programas corporativos de estágio.
type SelectionProgramSource struct {
	name       string
	feedURL    string
	client     *http.Client
	staticData []ProgramItem
}

// NewSelectionProgramSource cria a fonte de programas corporativos e processos seletivos de tecnologia.
func NewSelectionProgramSource(name, feedURL string, client *http.Client) *SelectionProgramSource {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return &SelectionProgramSource{
		name:    name,
		feedURL: feedURL,
		client:  client,
	}
}

// SetStaticData permite injetar itens estáticos (útil para fontes curadas ou testes offline).
func (s *SelectionProgramSource) SetStaticData(items []ProgramItem) {
	s.staticData = items
}

// Name retorna o identificador da fonte.
func (s *SelectionProgramSource) Name() string {
	return s.name
}

// Fetch consulta o feed ou os dados estáticos curados.
func (s *SelectionProgramSource) Fetch(ctx context.Context) ([]domain.RawJob, error) {
	if len(s.staticData) > 0 {
		return s.mapItemsToRaw(s.staticData), nil
	}

	if s.feedURL == "" {
		return []domain.RawJob{}, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição de feed: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "vagas-radar/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("falha ao consultar feed %s: %w", s.name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("feed retornou status não-200: %d", resp.StatusCode)
	}

	var items []ProgramItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("falha ao decodificar feed JSON: %w", err)
	}

	return s.mapItemsToRaw(items), nil
}

func (s *SelectionProgramSource) mapItemsToRaw(items []ProgramItem) []domain.RawJob {
	results := make([]domain.RawJob, 0, len(items))
	for _, item := range items {
		pub := item.PublishedAt
		if pub.IsZero() {
			pub = time.Now()
		}

		results = append(results, domain.RawJob{
			ExternalID:  item.ID,
			Title:       item.Title,
			Company:     item.Company,
			Description: item.Description,
			URL:         item.URL,
			LocationRaw: item.Location,
			WorkModeRaw: item.WorkMode,
			Source:      s.name,
			PublishedAt: pub,
		})
	}
	return results
}

// SetFeedURL altera a URL do feed (para testes unitários).
func (s *SelectionProgramSource) SetFeedURL(url string) {
	s.feedURL = strings.TrimSpace(url)
}
