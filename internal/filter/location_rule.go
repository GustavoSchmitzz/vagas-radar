package filter

import (
	"regexp"
	"strings"

	"github.com/GustavoSchmitzz/vagas-radar/internal/config"
	"github.com/GustavoSchmitzz/vagas-radar/internal/domain"
)

var (
	// Padrões de modalidades
	remotePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(remoto|remote|home\s*office|100%\s*remoto|teletrabalho)\b`),
	}
	hybridPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(hibrido|híbrido|hybrid)\b`),
	}
	onSitePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(presencial|on-site|onsite|no\s*local)\b`),
	}

	// Padrões de restrições geográficas excludentes em vagas declaradas remotas
	// Exemplo: "remoto apenas para residentes em SP", "exclusivo para moradores do Rio de Janeiro", etc.
	remoteStateExclusions = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:apenas|exclusiv[oa]|somente|obrigat[oó]ri[oa]|necess[aá]ri[oa]|residir|morar)\s+(?:para\s+)?(?:residentes\s+|moradores\s+)?(?:d[eo]\s+|em\s+)?(?:s[aã]o\s+paulo|sp|rio\s+de\s+janeiro|rj|curitiba|pr|parana|paran[aá]|minas\s+gerais|mg|belo\s+horizonte|bh|porto\s+alegre|rs|rio\s+grande\s+do\s+sul|santa\s+catarina|sc|florian[oó]polis|bras[ií]lia|df|bahia|ba|salvador|pernambuco|pe|recife|cear[aá]|ce|fortaleza)\b`),
		regexp.MustCompile(`(?i)\bresidentes?\s+(?:em|no|na)\s+(?:s[aã]o\s+paulo|sp|rj|pr|mg|rs|sc|df|ba|pe|ce)\b`),
		regexp.MustCompile(`(?i)\bvaga\s+remota\s+(?:apenas|somente)\s+para\b`),
	}
)

// LocationRule valida se a modalidade e a localização cumprem os critérios geográficos.
type LocationRule struct {
	cfg config.FilterConfig
}

// NewLocationRule instancia a regra de validação geográfica.
func NewLocationRule(cfg config.FilterConfig) *LocationRule {
	return &LocationRule{cfg: cfg}
}

// LocationEvaluationResult contém o resultado da checagem geográfica.
type LocationEvaluationResult struct {
	Location domain.Location
	WorkMode domain.WorkMode
	Accepted bool
	Reason   string
}

// Evaluate analisa a modalidade de trabalho e localização da vaga bruta.
func (r *LocationRule) Evaluate(raw domain.RawJob) LocationEvaluationResult {
	workMode := r.detectWorkMode(raw)
	loc := r.extractLocation(raw)

	switch workMode {
	case domain.WorkModeRemote:
		// Verificar se há restrições regionais no Brasil que excluam quem reside em MT
		fullText := strings.ToLower(raw.Title + " " + raw.LocationRaw + " " + raw.Description)
		for _, re := range remoteStateExclusions {
			// Se mencionar a restrição e NÃO incluir Cuiabá/MT
			if re.MatchString(fullText) && !strings.Contains(fullText, "cuiab") && !strings.Contains(fullText, "mato grosso") && !strings.Contains(fullText, " mt") {
				return LocationEvaluationResult{
					Location: loc,
					WorkMode: workMode,
					Accepted: false,
					Reason:   "Rejeitada: vaga remota com restrição geográfica para outro estado",
				}
			}
		}

		return LocationEvaluationResult{
			Location: loc,
			WorkMode: domain.WorkModeRemote,
			Accepted: true,
			Reason:   "Aprovada: vaga remota elegível para candidatos de qualquer estado do Brasil",
		}

	case domain.WorkModeOnSite, domain.WorkModeHybrid:
		// Vagas presenciais ou híbridas DEVEM ser em Cuiabá ou Várzea Grande - MT
		if r.isTargetCity(raw.LocationRaw) || r.isTargetCity(raw.Title) || r.isTargetCity(raw.Description) {
			modeStr := "presencial"
			if workMode == domain.WorkModeHybrid {
				modeStr = "híbrida"
			}
			return LocationEvaluationResult{
				Location: loc,
				WorkMode: workMode,
				Accepted: true,
				Reason:   "Aprovada: vaga " + modeStr + " em Cuiabá / Várzea Grande - MT",
			}
		}

		return LocationEvaluationResult{
			Location: loc,
			WorkMode: workMode,
			Accepted: false,
			Reason:   "Rejeitada: vaga presencial/híbrida fora de Cuiabá ou Várzea Grande (MT)",
		}

	default:
		// Se o modo for desconhecido, checamos se o texto cita Cuiabá/VG ou Remoto
		combined := strings.ToLower(raw.Title + " " + raw.LocationRaw + " " + raw.Description)
		if r.isTargetCity(combined) {
			return LocationEvaluationResult{
				Location: loc,
				WorkMode: domain.WorkModeOnSite,
				Accepted: true,
				Reason:   "Aprovada: localidade identificada em Cuiabá/Várzea Grande - MT",
			}
		}
		if matchesAny(combined, remotePatterns) {
			return LocationEvaluationResult{
				Location: loc,
				WorkMode: domain.WorkModeRemote,
				Accepted: true,
				Reason:   "Aprovada: modalidade remota inferida pelo texto",
			}
		}

		return LocationEvaluationResult{
			Location: loc,
			WorkMode: domain.WorkModeUnknown,
			Accepted: false,
			Reason:   "Rejeitada: localização/modalidade não compatível",
		}
	}
}

func (r *LocationRule) detectWorkMode(raw domain.RawJob) domain.WorkMode {
	candidate := strings.ToLower(raw.WorkModeRaw + " " + raw.LocationRaw + " " + raw.Title)

	// Ordem de precedência: híbrido antes de remoto puro
	if matchesAny(candidate, hybridPatterns) {
		return domain.WorkModeHybrid
	}
	if matchesAny(candidate, remotePatterns) {
		return domain.WorkModeRemote
	}
	if matchesAny(candidate, onSitePatterns) {
		return domain.WorkModeOnSite
	}

	return domain.WorkModeUnknown
}

func (r *LocationRule) isTargetCity(text string) bool {
	normalized := strings.ToLower(text)
	for _, city := range r.cfg.TargetCities {
		if strings.Contains(normalized, city) {
			return true
		}
	}
	return false
}

func (r *LocationRule) extractLocation(raw domain.RawJob) domain.Location {
	text := strings.ToLower(raw.LocationRaw + " " + raw.Title)
	loc := domain.Location{
		Raw:     raw.LocationRaw,
		Country: "Brasil",
	}

	if strings.Contains(text, "cuiab") {
		loc.City = "Cuiabá"
		loc.State = "MT"
	} else if strings.Contains(text, "varzea grande") || strings.Contains(text, "várzea grande") {
		loc.City = "Várzea Grande"
		loc.State = "MT"
	}

	return loc
}
