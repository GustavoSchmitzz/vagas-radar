package source

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/GustavoSchmitzz/vagas-radar/internal/domain"
)

type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title string    `xml:"title"`
	Items []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

// RSSSource coleta oportunidades a partir de feeds RSS de portais de tecnologia.
type RSSSource struct {
	name    string
	feedURL string
	client  *http.Client
}

// NewRSSSource instancia uma nova fonte RSS padrão.
func NewRSSSource(name, feedURL string, client *http.Client) *RSSSource {
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	return &RSSSource{
		name:    name,
		feedURL: feedURL,
		client:  client,
	}
}

// Name retorna o identificador da fonte.
func (s *RSSSource) Name() string {
	return s.name
}

// Fetch realiza a leitura e parse do XML do feed RSS.
func (s *RSSSource) Fetch(ctx context.Context) ([]domain.RawJob, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição para feed RSS %s: %w", s.name, err)
	}

	req.Header.Set("User-Agent", "vagas-radar/1.0 (Mozilla/5.0)")
	req.Header.Set("Accept", "application/rss+xml, application/xml, text/xml")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("falha ao consultar feed RSS %s: %w", s.name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("feed RSS %s retornou status %d", s.name, resp.StatusCode)
	}

	var feed rssFeed
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("falha ao decodificar XML do feed %s: %w", s.name, err)
	}

	rawJobs := make([]domain.RawJob, 0, len(feed.Channel.Items))
	for _, item := range feed.Channel.Items {
		cleanLink := strings.TrimSpace(item.Link)
		if cleanLink == "" {
			cleanLink = strings.TrimSpace(item.GUID)
		}

		pubDate, _ := time.Parse(time.RFC1123, item.PubDate)
		if pubDate.IsZero() {
			pubDate, _ = time.Parse(time.RFC1123Z, item.PubDate)
		}
		if pubDate.IsZero() {
			pubDate = time.Now()
		}

		company, cleanTitle := s.extractCompany(item.Title)

		rawJobs = append(rawJobs, domain.RawJob{
			ExternalID:  cleanLink,
			Title:       cleanTitle,
			Company:     company,
			Description: item.Description,
			URL:         cleanLink,
			LocationRaw: "Remoto",
			WorkModeRaw: "Remoto",
			Source:      s.name,
			PublishedAt: pubDate,
		})
	}

	return rawJobs, nil
}

func (s *RSSSource) extractCompany(title string) (company, cleanTitle string) {
	cleanTitle = strings.TrimSpace(title)
	company = "Confidencial / Empresa Tech"

	// Padrões comuns: "Company: Job Title" ou "Job Title at Company"
	if idx := strings.Index(cleanTitle, ": "); idx != -1 {
		company = strings.TrimSpace(cleanTitle[:idx])
		cleanTitle = strings.TrimSpace(cleanTitle[idx+2:])
	} else if idx := strings.LastIndex(strings.ToLower(cleanTitle), " at "); idx != -1 {
		company = strings.TrimSpace(cleanTitle[idx+4:])
	}

	return company, cleanTitle
}
